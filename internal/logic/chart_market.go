package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/biya-coin/injective-chronos-go/internal/cache"
	"github.com/biya-coin/injective-chronos-go/internal/model"
)

func (l *ChartLogic) getMarketHistoryByMarketIDs(ctx context.Context, marketIDs []string, resolution string, countback int) ([]model.MarketHistory, error) {
	var result []model.MarketHistory
	for _, mid := range marketIDs {
		// find latest candles for mid
		findOpts := options.Find()
		findOpts.SetSort(bson.D{{Key: "t", Value: -1}})
		if countback > 0 {
			findOpts.SetLimit(int64(countback))
		}
		cur, err := l.svcCtx.MarketColl.Find(ctx, bson.M{
			"kind":       "history",
			"market":     mid,
			"resolution": resolution,
		}, findOpts)
		if err != nil {
			return nil, err
		}
		var points []bson.M
		if err := cur.All(ctx, &points); err != nil {
			return nil, err
		}
		var out model.MarketHistory
		out.MarketID = mid
		out.Resolution = resolution
		for _, p := range points {
			data := p["data"].(bson.M)
			out.T = append(out.T, data["t"].(int64))
			out.O = append(out.O, data["o"].(float64))
			out.H = append(out.H, data["h"].(float64))
			out.L = append(out.L, data["l"].(float64))
			out.C = append(out.C, data["c"].(float64))
			out.V = append(out.V, data["v"].(float64))
		}
		result = append(result, out)
	}
	return result, nil
}

// GetMarketHistory returns the latest N candles per market from Mongo, aggregated by marketId.
// It reads documents inserted by cron_market, where each doc is one candle point.
func (l *ChartLogic) GetMarketHistory(ctx context.Context, marketIDs []string, resolution string, countback int) ([]model.MarketHistory, error) {
	if len(marketIDs) == 0 {
		return nil, fmt.Errorf("empty marketIDs")
	}
	if resolution == "" {
		resolution = "1"
	}
	// 把resolution转换为int除以2作为redis的baseTTLSeconds
	resolutionInt, _ := strconv.Atoi(resolution)
	baseTTLSeconds := resolutionInt * 60 / 2

	if countback <= 0 {
		countback = 0
	}

	cacheKey := fmt.Sprintf("chart:market:history:%s:%d:%v", resolution, countback, marketIDs)
	if bytes, err := cache.GetOrLoadBytes(
		ctx,
		l.svcCtx.Redis,
		cacheKey,
		baseTTLSeconds,
		1,
		l.svcCtx.Config.Redis.LockTTLSeconds,
		l.svcCtx.Config.Redis.RetryMs,
		l.svcCtx.Config.Redis.RetryMax,
		func(ctx context.Context) ([]byte, error) {
			result, err := l.getMarketHistoryByMarketIDs(ctx, marketIDs, resolution, countback)
			if err != nil {
				return nil, err
			}
			return json.Marshal(result)
		},
	); err == nil && bytes != nil {
		var v []model.MarketHistory
		if e := json.Unmarshal(bytes, &v); e == nil {
			return v, nil
		}
	}

	// fallback (no cache)
	result, err := l.getMarketHistoryByMarketIDs(ctx, marketIDs, resolution, countback)
	if err != nil {
		return nil, err
	}
	return result, nil
}
