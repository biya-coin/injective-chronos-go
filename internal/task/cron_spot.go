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
	var from int64
	var to int64 = time.Now().Unix()
	for _, res := range consts.SupportedMarketResolutions {
		// 动态计算 countback
		opts := options.FindOne().SetSort(bson.D{{Key: "t", Value: -1}}).SetProjection(bson.M{"data": 1})
		// 强类型解码，直接将 data 映射为 model.SpotMarketHistoryRaw，避免二次序列化
		var doc model.SpotHistoryDoc
		if err := svcCtx.SpotColl.FindOne(context.Background(), bson.M{"kind": "history", "resolution": res}, opts).Decode(&doc); err != nil {
			countback = 0
			from = 0
		} else {
			logx.Infof("fetchAndStoreSpotMarketHistory -> res %s lastT:%d", res, doc.Data.T)
			resolution, _ := strconv.ParseInt(res, 10, 64)
			countback = int((to-doc.Data.T)/(resolution*60)) + 5
			from = doc.Data.T
		}

		// 仅获取现货 marketIds（来源于 summary_all 快照）
		marketIDs := getMarketSummaryAllIds(svcCtx, "24h", consts.MarketTypeSpot)
		if len(marketIDs) == 0 {
			logx.Errorf("spot market history -> res %s: no market ids", res)
			continue
		}

		for _, mid := range marketIDs {

			protect("spot.market.history.batch", func() {
				rows, err := client.SpotMarketHistory(ctxBg, from, to, mid, res, countback)
				if err != nil {
					logx.Errorf("fetch spot market history -> res:%s market:%s: %v", res, mid, err)
					return
				}
				for tIndex := 0; tIndex < len(rows.T); tIndex++ {
					filter := bson.M{
						"kind":       "history",
						"market":     mid,
						"resolution": res,
						"t":          rows.T[tIndex],
					}
					count, err := svcCtx.SpotColl.CountDocuments(ctxBg, filter)
					if err != nil {
						logx.Errorf("count spot market history %s@%s: %v", mid, res, err)
						continue
					}
					if count == 0 {
						_, e := svcCtx.SpotColl.InsertOne(ctxBg, bson.M{
							"kind":       "history",
							"market":     mid,
							"resolution": res,
							"data": model.SpotMarketHistoryRaw{
								T: rows.T[tIndex],
								O: rows.O[tIndex],
								H: rows.H[tIndex],
								L: rows.L[tIndex],
								C: rows.C[tIndex],
								V: rows.V[tIndex],
							},
							"t":          rows.T[tIndex],
							"updated_at": time.Now(),
						})
						if e != nil {
							logx.Errorf("insert spot market history %s@%s: %v", mid, res, e)
						}
					}
				}
			})
		}
	}
}
