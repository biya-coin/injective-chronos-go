package task

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/biya-coin/injective-chronos-go/internal/consts"
	"github.com/biya-coin/injective-chronos-go/internal/injective"
	"github.com/biya-coin/injective-chronos-go/internal/model"
	"github.com/biya-coin/injective-chronos-go/internal/svc"
)

func fetchAndStoreSpotSummaryAll(ctxBg context.Context, svcCtx *svc.ServiceContext, client *injective.Client) {
	for _, res := range consts.SupportedResolutions {
		v, err := client.SpotMarketSummaryAll(ctxBg, res)
		if err != nil {
			logx.Errorf("fetch spot summary_all -> resolution %s: error %v", res, err)
			continue
		}
		_, e := svcCtx.SpotColl.InsertOne(ctxBg, bson.M{
			"kind":       "summary_all",
			"resolution": res,
			"data":       v,
			"updated_at": time.Now(),
		})
		if e != nil {
			logx.Errorf("insert spot summary_all -> resolution %s: error %v", res, e)
		}
	}
}

// fetchAndStoreSpotSummaries fetches per-spot market summary concurrently (bounded) and stores to Mongo.
func fetchAndStoreSpotSummaries(ctxBg context.Context, svcCtx *svc.ServiceContext, client *injective.Client) {
	for _, res := range consts.SupportedResolutions {
		marketIds := getMarketSummaryAllIds(svcCtx, res, consts.MarketTypeSpot)
		if marketIds == nil {
			logx.Errorf("get market spot summary all ids -> resolution %s: error %v", res, marketIds)
			continue
		}
		// bounded concurrency
		const maxWorkers = 8
		sem := make(chan struct{}, maxWorkers)
		var wg sync.WaitGroup
		for _, mid := range marketIds {
			sem <- struct{}{}
			wg.Add(1)
			go func(mid string) {
				defer wg.Done()
				defer func() { <-sem }()
				defer recoverAndLog("spot.worker:" + res + ":" + mid)
				one, err := client.SpotMarketSummaryAtResolution(ctxBg, mid, res)
				if err != nil {
					logx.Errorf("fetch spot summary %s@%s: %v", mid, res, err)
					return
				}
				_, e := svcCtx.SpotColl.InsertOne(ctxBg, bson.M{
					"kind":       "summary",
					"market":     mid,
					"resolution": res,
					"data":       *one,
					"updated_at": time.Now(),
				})
				if e != nil {
					logx.Errorf("insert spot summary %s@%s: %v", mid, res, e)
				}
			}(mid)
		}
		wg.Wait()
	}
}

// fetchAndStoreSpotConfig fetches spot config from Injective and stores to Mongo.
func fetchAndStoreSpotConfig(ctxBg context.Context, svcCtx *svc.ServiceContext, client *injective.Client) {
	cfg, err := client.SpotConfig(ctxBg)
	if err != nil {
		logx.Errorf("fetch spot config: %v", err)
		return
	}
	_, e := svcCtx.SpotColl.InsertOne(ctxBg, bson.M{
		"kind":       "config",
		"data":       cfg,
		"updated_at": time.Now(),
	})
	if e != nil {
		logx.Errorf("insert spot config: %v", e)
	}
}

// fetchAndStoreSpotMarketHistory fetches spot-only market history and stores into Mongo `MarketColl`.
func fetchAndStoreSpotMarketHistory(ctxBg context.Context, svcCtx *svc.ServiceContext, client *injective.Client) {
	var countback = 10
	for _, res := range consts.SupportedMarketResolutions {
		// 动态计算 countback
		opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
		var doc bson.M
		if err := svcCtx.MarketColl.FindOne(context.Background(), bson.M{"kind": "history", "resolution": res}, opts).Decode(&doc); err != nil {
			countback = 0
		} else {
			resolution, _ := strconv.ParseInt(res, 10, 64)
			countback = int((time.Now().Unix()-doc["data"].(model.MarketHistoryRaw).T)/resolution) + 10
		}

		// 仅获取现货 marketIds（来源于 summary_all 快照）
		marketIDs := getMarketSummaryAllIds(svcCtx, "24h", consts.MarketTypeSpot)
		if len(marketIDs) == 0 {
			logx.Errorf("spot market history -> res %s: no market ids", res)
			continue
		}

		const batchSize = 50
		for i := 0; i < len(marketIDs); i += batchSize {
			end := i + batchSize
			if end > len(marketIDs) {
				end = len(marketIDs)
			}
			batch := marketIDs[i:end]
			protect("spot.market.history.batch", func() {
				rows, err := client.MarketHistory(ctxBg, batch, res, countback)
				if err != nil {
					logx.Errorf("fetch spot market history -> res %s batch %d-%d: %v", res, i, end, err)
					return
				}
				for _, row := range rows {
					for tIndex := 0; tIndex < len(row.T); tIndex++ {
						filter := bson.M{
							"kind":       "history",
							"market":     row.MarketID,
							"resolution": res,
							"t":          row.T[tIndex],
						}
						count, err := svcCtx.MarketColl.CountDocuments(ctxBg, filter)
						if err != nil {
							logx.Errorf("count spot market history %s@%s: %v", row.MarketID, res, err)
							continue
						}
						if count == 0 {
							_, e := svcCtx.MarketColl.InsertOne(ctxBg, bson.M{
								"kind":       "history",
								"market":     row.MarketID,
								"resolution": res,
								"data": model.MarketHistoryRaw{
									MarketID:   row.MarketID,
									Resolution: res,
									T:          row.T[tIndex],
									O:          row.O[tIndex],
									H:          row.H[tIndex],
									L:          row.L[tIndex],
									C:          row.C[tIndex],
									V:          row.V[tIndex],
								},
								"t":          row.T[tIndex],
								"updated_at": time.Now(),
							})
							if e != nil {
								logx.Errorf("insert spot market history %s@%s: %v", row.MarketID, res, e)
							}
						}
					}
				}
			})
		}
	}
}
