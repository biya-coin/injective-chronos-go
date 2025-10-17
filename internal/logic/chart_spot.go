package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/biya-coin/injective-chronos-go/internal/cache"
	"github.com/biya-coin/injective-chronos-go/internal/consts"
	"github.com/biya-coin/injective-chronos-go/internal/model"
	"github.com/zeromicro/go-zero/core/logx"
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

func (l *ChartLogic) getMarketHistorySpotByMarketIDs(ctx context.Context, marketId string, resolution string, countback int, from int64, to int64) (model.SpotMarketHistory, error) {
	findOpts := options.Find()
	findOpts.SetSort(bson.D{{Key: "t", Value: -1}})
	if countback > 0 {
		findOpts.SetLimit(int64(countback))
	}
	cur, err := l.svcCtx.SpotColl.Find(ctx, bson.M{
		"kind":       "history",
		"market":     marketId,
		"resolution": resolution,
		"t":          bson.M{"$gte": from, "$lte": to},
	}, findOpts)
	if err != nil {
		logx.Errorf("getMarketHistorySpotByMarketIDs find error: %v", err)
		return model.SpotMarketHistory{}, err
	}
	var points []model.SpotHistoryDoc
	if err := cur.All(ctx, &points); err != nil {
		logx.Errorf("getMarketHistorySpotByMarketIDs all error: %v", err)
		return model.SpotMarketHistory{}, err
	}
	var out model.SpotMarketHistory = model.SpotMarketHistory{
		T: make([]int64, 0),
		O: make([]float64, 0),
		H: make([]float64, 0),
		L: make([]float64, 0),
		C: make([]float64, 0),
		V: make([]float64, 0),
	}
	logx.Debugf("getMarketHistorySpotByMarketIDs----------------> points: %+v", points)
	for _, p := range points {
		data := p.Data
		// 使用bson unmarshal
		out.T = append(out.T, data.T)
		out.O = append(out.O, data.O)
		out.H = append(out.H, data.H)
		out.L = append(out.L, data.L)
		out.C = append(out.C, data.C)
		out.V = append(out.V, data.V)
	}
	return out, nil
}

func (l *ChartLogic) GetMarketHistorySpot(ctx context.Context, marketId string, resolution string, countback int, from int64, to int64) (model.SpotMarketHistory, error) {
	cacheKey := fmt.Sprintf("chart:spot:history:%s:%d:%d:%d:%s", resolution, countback, from, to, marketId)
	if marketId == "" {
		return model.SpotMarketHistory{}, fmt.Errorf("empty marketId")
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
			result, err := l.getMarketHistorySpotByMarketIDs(ctx, marketId, resolution, countback, from, to)
			if err != nil {
				logx.Errorf("getMarketHistorySpotByMarketIDs error: %v", err)
				return nil, err
			}
			return json.Marshal(result)
		},
	); err == nil && bytes != nil {
		var v model.SpotMarketHistory
		if e := json.Unmarshal(bytes, &v); e == nil {
			return v, nil
		}
	}

	// fallback (no cache)
	result, err := l.getMarketHistorySpotByMarketIDs(ctx, marketId, resolution, countback, from, to)
	if err != nil {
		logx.Errorf("getMarketHistorySpotByMarketIDs error: %v", err)
		return model.SpotMarketHistory{}, err
	}
	return result, nil
}

func (l *ChartLogic) getSpotSymbolInfoFromDB(ctx context.Context, group string) (*model.SpotSymbolInfo, error) {
	cur, err := l.svcCtx.SpotColl.Find(ctx, bson.M{"kind": "symbol_info", "group": group})
	if err != nil {
		return nil, err
	}
	var doc []model.SpotSymbolInfoRawDoc
	if err := cur.All(ctx, &doc); err != nil {
		return nil, err
	}
	if len(doc) == 0 {
		return nil, fmt.Errorf("no symbol info found")
	}
	IntradayMultipliers := doc[0].Data.IntradayMultipliers
	var out model.SpotSymbolInfo = model.SpotSymbolInfo{
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

func (l *ChartLogic) GetSpotSymbolInfo(ctx context.Context, group string) (*model.SpotSymbolInfo, error) {
	cacheKey := fmt.Sprintf("chart:spot:symbol_info:%s", group)
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
			doc, err := l.getSpotSymbolInfoFromDB(ctx, group)
			if err != nil {
				logx.Errorf("getSpotSymbolInfoFromDB error: %v", err)
				return nil, err
			}
			return json.Marshal(doc)
		},
	); err == nil && bytes != nil {
		var v model.SpotSymbolInfo
		if e := json.Unmarshal(bytes, &v); e == nil {
			return &v, nil
		}
	}
	doc, err := l.getSpotSymbolInfoFromDB(ctx, group)
	if err != nil {
		logx.Errorf("getSpotSymbolInfoFromDB error: %v", err)
		return nil, err
	}
	return doc, nil
}

func (l *ChartLogic) getSpotSymbolsFromDB(ctx context.Context, symbol string) (*model.SpotSymbolsRaw, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	var doc model.SpotSymbolsRawDoc
	if err := l.svcCtx.SpotColl.FindOne(ctx, bson.M{"kind": "symbols", "symbol": symbol}, opts).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc.Data, nil
}

func (l *ChartLogic) GetSpotSymbols(ctx context.Context, symbol string) (*model.SpotSymbolsRaw, error) {

	cacheKey := fmt.Sprintf("chart:spot:symbols:%s", symbol)
	if symbol == "" {
		return nil, fmt.Errorf("empty symbol")
	}
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
			doc, err := l.getSpotSymbolsFromDB(ctx, symbol)
			if err != nil {
				logx.Errorf("getSpotSymbolsFromDB error: %v", err)
				return nil, err
			}
			return json.Marshal(doc)
		},
	); err == nil && bytes != nil {
		var v model.SpotSymbolsRaw
		if e := json.Unmarshal(bytes, &v); e == nil {
			return &v, nil
		}
	}
	doc, err := l.getSpotSymbolsFromDB(ctx, symbol)
	if err != nil {
		logx.Errorf("getSpotSymbolsFromDB error: %v", err)
		return nil, err
	}
	return doc, nil

}
