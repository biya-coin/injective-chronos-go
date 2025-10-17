package task

import (
	"context"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/biya-coin/injective-chronos-go/internal/consts"
	"github.com/biya-coin/injective-chronos-go/internal/injective"
	"github.com/biya-coin/injective-chronos-go/internal/model"
	"github.com/biya-coin/injective-chronos-go/internal/svc"
)

func getMarketHistoryAllIds(svcCtx *svc.ServiceContext, resolution string) []string {
	// get market ids from summary_all snapshots for the configured resolution
	spotIds := getMarketSummaryAllIds(svcCtx, resolution, consts.MarketTypeSpot)
	derivativeIds := getMarketSummaryAllIds(svcCtx, resolution, consts.MarketTypeDerivative)
	if len(spotIds) == 0 && len(derivativeIds) == 0 {
		cronErrorf("market history -> resolution %s: no market ids from summary_all", resolution)
		return nil
	}
	// deduplicate
	uniq := make(map[string]struct{})
	for _, sid := range spotIds {
		if sid != "" {
			uniq[sid] = struct{}{}
		}
	}
	for _, did := range derivativeIds {
		if did != "" {
			uniq[did] = struct{}{}
		}
	}
	var marketIDs []string
	for mid := range uniq {
		marketIDs = append(marketIDs, mid)
	}
	if len(marketIDs) == 0 {
		cronErrorf("market history -> resolution %s: merged market ids empty", resolution)
		return nil
	}
	return marketIDs
}

// fetchAndStoreMarketHistory aggregates market IDs from spot and derivative summary_all, then fetches
// market candle history in batches and stores records to Mongo `MarketColl`.
func fetchAndStoreMarketHistory(ctxBg context.Context, svcCtx *svc.ServiceContext, client *injective.Client) {
	var countback = 0
	for _, res := range consts.SupportedMarketResolutions {

		// 这里需要动态去算最新改的countback
		// 先查数据库是否有数据，如果没有，那countback就是0
		opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
		var doc model.MarketHistoryRawDoc
		if err := svcCtx.MarketColl.FindOne(context.Background(), bson.M{"kind": "history", "resolution": res}, opts).Decode(&doc); err != nil {
			countback = 0
		} else {
			// 当前时间的时间戳减去最新的一条数据的timestamp，得到时间差，再除以resolution，得到countback
			// 加10是为了防止数据不足，导致countback为0
			resolution, _ := strconv.ParseInt(res, 10, 64)
			countback = int((time.Now().Unix()-doc.Data.T)/resolution) + 10
		}
		if countback > 1440 {
			countback = 1440
		}
		marketIDs := getMarketHistoryAllIds(svcCtx, "24h")
		// batch request to avoid very long query string
		for _, marketId := range marketIDs {
			protect("market.history.batch", func() {
				rows, err := client.MarketHistory(ctxBg, []string{marketId}, res, countback)
				if err != nil {
					cronErrorf("fetch market history -> res %s marketId %s: %v", res, marketId, err)
					return
				}
				for _, row := range rows {
					for t_index := 0; t_index < len(row.T); t_index++ {
						// 先查询是否已存在该条记录，不存在则插入
						filter := bson.M{
							"kind":       "history",
							"marketId":   row.MarketID,
							"resolution": res,
							"t":          row.T[t_index],
						}
						count, err := svcCtx.MarketColl.CountDocuments(ctxBg, filter)
						if err != nil {
							cronErrorf("count market history %s@%s: %v", row.MarketID, res, err)
							continue
						}
						if count == 0 {
							_, e := svcCtx.MarketColl.InsertOne(ctxBg, bson.M{
								"kind":       "history",
								"marketId":   row.MarketID,
								"resolution": res,
								"data": model.MarketHistoryRaw{
									MarketID:   row.MarketID,
									Resolution: res,
									T:          row.T[t_index],
									O:          row.O[t_index],
									H:          row.H[t_index],
									L:          row.L[t_index],
									C:          row.C[t_index],
									V:          row.V[t_index],
								},
								"t":          row.T[t_index],
								"updated_at": time.Now(),
							})
							if e != nil {
								cronErrorf("insert market history %s@%s: %v", row.MarketID, res, e)
							}
						}
					}
				}
			})
		}
	}
}
