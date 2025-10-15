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
	"github.com/biya-coin/injective-chronos-go/internal/injective"
	"github.com/biya-coin/injective-chronos-go/internal/model"
)

// getMarketSummaryAllDerivative returns the latest derivative summary_all, with Redis caching.
func (l *ChartLogic) getMarketSummaryAllDerivative(ctx context.Context, resolution string) ([]model.DerivativeMarketSummary, error) {
	cacheKey := fmt.Sprintf("chart:summary_all:%s", consts.MarketTypeDerivative)
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
			if err := l.svcCtx.DerivativeColl.FindOne(ctx, bson.M{"kind": "summary_all", "resolution": resolution}, opts).Decode(&doc); err != nil {
				return nil, err
			}
			return json.Marshal(doc["data"])
		},
	); err == nil && bytes != nil {
		var v []model.DerivativeMarketSummary
		if e := json.Unmarshal(bytes, &v); e == nil {
			return v, nil
		}
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	var doc bson.M
	if err := l.svcCtx.DerivativeColl.FindOne(ctx, bson.M{"kind": "summary_all", "resolution": resolution}, opts).Decode(&doc); err != nil {
		return nil, err
	}
	bytes, _ := json.Marshal(doc["data"])
	_ = l.svcCtx.Redis.Set(ctx, cacheKey, bytes, 5*time.Minute).Err()
	var v []model.DerivativeMarketSummary
	err := json.Unmarshal(bytes, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// getMarketSummaryDerivative returns a single derivative market summary using resolution.
func (l *ChartLogic) getMarketSummaryDerivative(ctx context.Context, market string, resolution string) (*model.DerivativeMarketSummary, error) {
	cacheKey := fmt.Sprintf("chart:summary:%s:%s:%s", consts.MarketTypeDerivative, resolution, market)
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
			if err := l.svcCtx.DerivativeColl.FindOne(ctx, filter, opts).Decode(&doc); err != nil {
				return nil, err
			}
			return json.Marshal(doc["data"])
		},
	); err == nil && bytes != nil {
		var v model.DerivativeMarketSummary
		err := json.Unmarshal(bytes, &v)
		if err != nil {
			return nil, err
		}
		return &v, nil
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	var doc bson.M
	filter := bson.M{"kind": "summary", "market": market, "resolution": resolution}
	if err := l.svcCtx.DerivativeColl.FindOne(ctx, filter, opts).Decode(&doc); err == nil {
		bytes, _ := json.Marshal(doc["data"]) // assume data field holds payload
		_ = l.svcCtx.Redis.Set(ctx, cacheKey, bytes, 5*time.Minute).Err()
		var v model.DerivativeMarketSummary
		err := json.Unmarshal(bytes, &v)
		if err != nil {
			return nil, err
		}
		return &v, nil
	}
	// not found; keep nil
	return nil, nil
}

// GetDerivativeConfig returns derivative TradingView-style config from Injective with Redis caching.
func (l *ChartLogic) GetDerivativeConfig(ctx context.Context) (*model.ChartDerivativeConfig, error) {
	cacheKey := "chart:derivative:config"
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
			client := injective.NewClient(l.svcCtx.Config.Injective, l.svcCtx.HttpClient)
			cfg, err := client.DerivativeConfig(ctx)
			if err != nil {
				return nil, err
			}
			return json.Marshal(cfg)
		},
	); err == nil && bytes != nil {
		var v model.ChartDerivativeConfig
		if e := json.Unmarshal(bytes, &v); e == nil {
			return &v, nil
		}
	}
	// fallback
	client := injective.NewClient(l.svcCtx.Config.Injective, l.svcCtx.HttpClient)
	cfg, err := client.DerivativeConfig(ctx)
	if err != nil {
		return nil, err
	}
	bytes, _ := json.Marshal(cfg)
	_ = l.svcCtx.Redis.Set(ctx, cacheKey, bytes, 5*time.Minute).Err()
	return cfg, nil
}
