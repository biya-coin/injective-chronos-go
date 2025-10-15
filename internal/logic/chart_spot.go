package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/biya-coin/injective-chronos-go/internal/cache"
	"github.com/biya-coin/injective-chronos-go/internal/consts"
	"github.com/biya-coin/injective-chronos-go/internal/model"
)

// GetSpotConfig returns spot TradingView-style config from Injective with Redis caching.
func (l *ChartLogic) GetSpotConfig(ctx context.Context) (*model.ChartSpotConfig, error) {
	cacheKey := "chart:spot:config"
	if bytes, err := cache.GetOrLoadBytes(
		ctx,
		l.svcCtx.Redis,
		cacheKey,
		l.svcCtx.Config.Redis.TTLSeconds,
		l.svcCtx.Config.Redis.JitterSeconds,
		l.svcCtx.Config.Redis.LockTTLSeconds,
		l.svcCtx.Config.Redis.RetryMs,
		l.svcCtx.Config.Redis.RetryMax,
		func(ctx context.Context) ([]byte, error) {
			opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
			var doc bson.M
			if err := l.svcCtx.SpotColl.FindOne(ctx, bson.M{"kind": "config"}, opts).Decode(&doc); err != nil {
				return nil, err
			}
			return json.Marshal(doc["data"])
		},
	); err == nil && bytes != nil {
		var v model.ChartSpotConfig
		if e := json.Unmarshal(bytes, &v); e == nil {
			return &v, nil
		}
	}
	// fallback（无缓存命中时直接读库）
	opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	var doc bson.M
	if err := l.svcCtx.SpotColl.FindOne(ctx, bson.M{"kind": "config"}, opts).Decode(&doc); err != nil {
		return nil, err
	}
	bytes, _ := json.Marshal(doc["data"])
	_ = l.svcCtx.Redis.Set(ctx, cacheKey, bytes, 5*time.Minute).Err()
	var v model.ChartSpotConfig
	if e := json.Unmarshal(bytes, &v); e != nil {
		return nil, e
	}
	return &v, nil
}

// getMarketSummaryAllSpot returns the latest spot summary_all, with Redis caching.
func (l *ChartLogic) getMarketSummaryAllSpot(ctx context.Context, resolution string) ([]model.SpotMarketSummary, error) {
	cacheKey := fmt.Sprintf("chart:summary_all:%s", consts.MarketTypeSpot)
	if bytes, err := cache.GetOrLoadBytes(
		ctx,
		l.svcCtx.Redis,
		cacheKey,
		l.svcCtx.Config.Redis.TTLSeconds,
		l.svcCtx.Config.Redis.JitterSeconds,
		l.svcCtx.Config.Redis.LockTTLSeconds,
		l.svcCtx.Config.Redis.RetryMs,
		l.svcCtx.Config.Redis.RetryMax,
		func(ctx context.Context) ([]byte, error) {
			opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
			var doc bson.M
			if err := l.svcCtx.SpotColl.FindOne(ctx, bson.M{"kind": "summary_all", "resolution": resolution}, opts).Decode(&doc); err != nil {
				return nil, err
			}
			return json.Marshal(doc["data"])
		},
	); err == nil && bytes != nil {
		var v []model.SpotMarketSummary
		if e := json.Unmarshal(bytes, &v); e == nil {
			return v, nil
		}
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	var doc bson.M
	if err := l.svcCtx.SpotColl.FindOne(ctx, bson.M{"kind": "summary_all", "resolution": resolution}, opts).Decode(&doc); err != nil {
		return nil, err
	}
	bytes, _ := json.Marshal(doc["data"])
	_ = l.svcCtx.Redis.Set(ctx, cacheKey, bytes, 5*time.Minute).Err()
	var v []model.SpotMarketSummary
	err := json.Unmarshal(bytes, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// getMarketSummarySpot returns a single spot market summary.
func (l *ChartLogic) getMarketSummarySpot(ctx context.Context, market string, resolution string) (*model.SpotMarketSummary, error) {
	cacheKey := fmt.Sprintf("chart:summary:%s:%s:%s", consts.MarketTypeSpot, resolution, market)
	if bytes, err := cache.GetOrLoadBytes(
		ctx,
		l.svcCtx.Redis,
		cacheKey,
		l.svcCtx.Config.Redis.TTLSeconds,
		l.svcCtx.Config.Redis.JitterSeconds,
		l.svcCtx.Config.Redis.LockTTLSeconds,
		l.svcCtx.Config.Redis.RetryMs,
		l.svcCtx.Config.Redis.RetryMax,
		func(ctx context.Context) ([]byte, error) {
			opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
			var doc bson.M
			filter := bson.M{"kind": "summary", "market": market, "resolution": resolution}
			if err := l.svcCtx.SpotColl.FindOne(ctx, filter, opts).Decode(&doc); err != nil {
				return nil, err
			}
			return json.Marshal(doc["data"])
		},
	); err == nil && bytes != nil {
		var v model.SpotMarketSummary
		err := json.Unmarshal(bytes, &v)
		if err != nil {
			return nil, err
		}
		return &v, nil
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	var doc bson.M
	filter := bson.M{"kind": "summary", "market": market, "resolution": resolution}
	if err := l.svcCtx.SpotColl.FindOne(ctx, filter, opts).Decode(&doc); err == nil {
		bytes, _ := json.Marshal(doc["data"]) // assume data field holds payload
		_ = l.svcCtx.Redis.Set(ctx, cacheKey, bytes, 5*time.Minute).Err()
		var v model.SpotMarketSummary
		err := json.Unmarshal(bytes, &v)
		if err != nil {
			return nil, err
		}
		return &v, nil
	}
	// not found; keep nil
	return nil, nil
}
