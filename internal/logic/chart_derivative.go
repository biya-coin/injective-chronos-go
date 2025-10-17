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

func (l *ChartLogic) getDerivativeSymbolInfoFromDB(ctx context.Context, group string) (*model.DerivativeSymbolInfo, error) {
	cur, err := l.svcCtx.DerivativeColl.Find(ctx, bson.M{"kind": "symbol_info", "group": group})
	if err != nil {
		return nil, err
	}
	var doc []model.DerivativeSymbolInfoRawDoc
	if err := cur.All(ctx, &doc); err != nil {
		return nil, err
	}
	if len(doc) == 0 {
		return nil, fmt.Errorf("no symbol info found")
	}
	IntradayMultipliers := doc[0].Data.IntradayMultipliers
	var out model.DerivativeSymbolInfo = model.DerivativeSymbolInfo{
		Symbol:              make([]string, 0),
		Name:                make([]string, 0),
		Description:         make([]string, 0),
		Currency:            make([]string, 0),
		ExchangeListed:      make([]string, 0),
		ExchangeTraded:      make([]string, 0),
		Minmovement:         make([]int, 0),
		Pricescale:          make([]int, 0),
		Timezone:            make([]string, 0),
		Type:                make([]string, 0),
		SessionRegular:      make([]string, 0),
		BaseCurrency:        make([]string, 0),
		HasIntraday:         make([]bool, 0),
		Ticker:              make([]string, 0),
		IntradayMultipliers: IntradayMultipliers,
		BarFillgaps:         make([]bool, 0),
	}
	for _, d := range doc {
		out.Symbol = append(out.Symbol, d.Data.Symbol)
		out.Name = append(out.Name, d.Data.Name)
		out.Description = append(out.Description, d.Data.Description)
		out.Currency = append(out.Currency, d.Data.Currency)
		out.ExchangeListed = append(out.ExchangeListed, d.Data.ExchangeListed)
		out.ExchangeTraded = append(out.ExchangeTraded, d.Data.ExchangeTraded)
		out.Minmovement = append(out.Minmovement, d.Data.Minmovement)
		out.Pricescale = append(out.Pricescale, d.Data.Pricescale)
		out.Timezone = append(out.Timezone, d.Data.Timezone)
		out.Type = append(out.Type, d.Data.Type)
		out.SessionRegular = append(out.SessionRegular, d.Data.SessionRegular)
		out.BaseCurrency = append(out.BaseCurrency, d.Data.BaseCurrency)
		out.HasIntraday = append(out.HasIntraday, d.Data.HasIntraday)
		out.Ticker = append(out.Ticker, d.Data.Ticker)
		out.BarFillgaps = append(out.BarFillgaps, d.Data.BarFillgaps)
	}
	return &out, nil
}

func (l *ChartLogic) GetDerivativeSymbolInfo(ctx context.Context, group string) (*model.DerivativeSymbolInfo, error) {
	cacheKey := fmt.Sprintf("chart:derivative:symbol_info:%s", group)
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
			symbolInfo, err := l.getDerivativeSymbolInfoFromDB(ctx, group)
			if err != nil {
				return nil, err
			}
			return json.Marshal(symbolInfo)
		},
	); err == nil && bytes != nil {
		var doc model.DerivativeSymbolInfo
		if e := json.Unmarshal(bytes, &doc); e == nil {
			return &doc, nil
		}
	}
	doc, err := l.getDerivativeSymbolInfoFromDB(ctx, group)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (l *ChartLogic) getDerivativeSymbolsFromDB(ctx context.Context, symbol string) (*model.DerivativeSymbolsRaw, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	var doc model.DerivativeSymbolsRawDoc
	if err := l.svcCtx.DerivativeColl.FindOne(ctx, bson.M{"kind": "symbols", "symbol": symbol}, opts).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc.Data, nil
}

func (l *ChartLogic) GetDerivativeSymbols(ctx context.Context, symbol string) (*model.DerivativeSymbolsRaw, error) {
	cacheKey := fmt.Sprintf("chart:derivative:symbols:%s", symbol)
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
			doc, err := l.getDerivativeSymbolsFromDB(ctx, symbol)
			if err != nil {
				return nil, err
			}
			return json.Marshal(doc)
		},
	); err == nil && bytes != nil {
		var v model.DerivativeSymbolsRaw
		if e := json.Unmarshal(bytes, &v); e == nil {
			return &v, nil
		}
	}
	doc, err := l.getDerivativeSymbolsFromDB(ctx, symbol)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
