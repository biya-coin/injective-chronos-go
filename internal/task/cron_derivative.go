package task

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/biya-coin/injective-chronos-go/internal/consts"
	"github.com/biya-coin/injective-chronos-go/internal/injective"
	"github.com/biya-coin/injective-chronos-go/internal/model"
	"github.com/biya-coin/injective-chronos-go/internal/svc"
)

func fetchAndStoreDerivativeConfig(ctxBg context.Context, svcCtx *svc.ServiceContext, client *injective.Client) {
	defer func() {
		if r := recover(); r != nil {
			cronErrorf("goroutine recovered from fetchAndStoreDerivativeConfig: %v", r)
		}
	}()
	if release, ok := acquireTaskLock(ctxBg, svcCtx, "derivative_config_fetch", 3*time.Second); !ok {
		cronInfof("fetchAndStoreDerivativeConfig: acquire lock timeout, skip this run")
		return
	} else {
		defer release()
	}

	cfg, err := client.DerivativeConfig(ctxBg)
	if err != nil {
		cronErrorf("fetch derivative config: %v", err)
		return
	}
	_, e := svcCtx.DerivativeColl.InsertOne(ctxBg, model.ChartDerivativeConfigRawDoc{
		Kind:      "config",
		UpdatedAt: time.Now(),
		Data:      *cfg,
	})
	if e != nil {
		cronErrorf("insert derivative config: %v", e)
	}
}
func fetchAndStoreDerivativeSummaryAll(ctxBg context.Context, svcCtx *svc.ServiceContext, client *injective.Client) {
	defer func() {
		if r := recover(); r != nil {
			cronErrorf("goroutine recovered from fetchAndStoreDerivativeSummaryAll: %v", r)
		}
	}()
	if release, ok := acquireTaskLock(ctxBg, svcCtx, "derivative_summary_all_fetch", 3*time.Second); !ok {
		cronInfof("fetchAndStoreDerivativeSummaryAll: acquire lock timeout, skip this run")
		return
	} else {
		defer release()
	}
	for _, res := range consts.SupportedResolutions {
		v, err := client.DerivativeMarketSummaryAll(ctxBg, res)
		if err != nil {
			cronErrorf("fetch derivative summary_all -> resolution %s: error %v", res, err)
			continue
		}
		_, e := svcCtx.DerivativeColl.InsertOne(ctxBg, bson.M{
			"kind":       "summary_all",
			"resolution": res,
			"data":       v,
			"updated_at": time.Now(),
		})
		if e != nil {
			cronErrorf("insert derivative summary_all -> resolution %s: error %v", res, e)
		}
	}
}

func fetchAndStoreDerivativeSummaries(ctxBg context.Context, svcCtx *svc.ServiceContext, client *injective.Client) {
	defer func() {
		if r := recover(); r != nil {
			cronErrorf("goroutine recovered from fetchAndStoreDerivativeSummaries: %v", r)
		}
	}()
	if release, ok := acquireTaskLock(ctxBg, svcCtx, "derivative_summaries_fetch", 3*time.Second); !ok {
		cronInfof("fetchAndStoreDerivativeSummaries: acquire lock timeout, skip this run")
		return
	} else {
		defer release()
	}
	for _, res := range consts.SupportedResolutions {
		marketIds := getMarketSummaryAllIds(svcCtx, res, consts.MarketTypeDerivative)
		if marketIds == nil {
			cronErrorf("get market derivative summary all ids -> resolution %s: error %v", res, marketIds)
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
				defer recoverAndLog("derivative.worker:" + res + ":" + mid)
				ctx, cancel := context.WithTimeout(ctxBg, 30*time.Second)
				defer cancel()
				one, err := client.DerivativeMarketSummaryAtResolution(ctx, mid, res)
				if err != nil {
					cronErrorf("fetch derivative summary %s %s: %v", mid, res, err)
					return
				}
				_, e := svcCtx.DerivativeColl.InsertOne(ctxBg, bson.M{
					"kind":       "summary",
					"market":     mid,
					"resolution": res,
					"data":       *one,
					"updated_at": time.Now(),
				})
				if e != nil {
					cronErrorf("insert derivative summary %s %s: %v", mid, res, e)
				}
			}(mid)
		}
		wg.Wait()
	}
}

func getDerivativeSymbolsList(ctxBg context.Context, svcCtx *svc.ServiceContext, group string) ([]string, error) {
	filter := bson.M{"kind": "symbol_info", "group": group}
	cur, err := svcCtx.DerivativeColl.Find(ctxBg, filter)
	if err != nil {
		cronErrorf("get derivative symbols list error: %v", err)
		return nil, err
	}
	var symbols []model.DerivativeSymbolInfoRawDoc
	var symbolList []string
	if err := cur.All(ctxBg, &symbols); err != nil {
		cronErrorf("get derivative symbols list error: %v", err)
		return nil, err
	}
	for _, symbol := range symbols {
		symbolList = append(symbolList, symbol.Symbol)
	}
	return symbolList, nil
}

func fetchAndStoreDerivativeSymbolInfo(ctxBg context.Context, svcCtx *svc.ServiceContext, client *injective.Client) {
	defer func() {
		if r := recover(); r != nil {
			cronErrorf("goroutine recovered from fetchAndStoreDerivativeSymbolInfo: %v", r)
		}
	}()
	if release, ok := acquireTaskLock(ctxBg, svcCtx, "derivative_symbol_info_fetch", 3*time.Second); !ok {
		cronInfof("fetchAndStoreDerivativeSymbolInfo: acquire lock timeout, skip this run")
		return
	} else {
		defer release()
	}

	// TOOD:to confirm the group come form
	var group string = ""
	drivativeSymbolInfo, err := client.DerivativeSymbolInfo(ctxBg, group)
	if err != nil {
		cronErrorf("fetch derivative symbol info -> group:%s: %v", group, err)
		return
	}
	for index := 0; index < len(drivativeSymbolInfo.Symbol); index++ {
		filter := bson.M{
			"kind":   "symbol_info",
			"symbol": drivativeSymbolInfo.Symbol[index],
			"group":  group,
		}
		count, err := svcCtx.DerivativeColl.CountDocuments(ctxBg, filter)
		if err != nil {
			cronErrorf("count derivative symbol info -> symbol:%s: %v", drivativeSymbolInfo.Symbol[index], err)
			continue
		}

		if count == 0 {
			// insert into mongo
			_, err = svcCtx.DerivativeColl.InsertOne(ctxBg, model.DerivativeSymbolInfoRawDoc{
				Kind:      "symbol_info",
				Symbol:    drivativeSymbolInfo.Symbol[index],
				Group:     group,
				UpdatedAt: time.Now(),
				Data: model.DerivativeSymbolInfoRaw{
					Symbol:              drivativeSymbolInfo.Symbol[index],
					Name:                drivativeSymbolInfo.Name[index],
					Description:         drivativeSymbolInfo.Description[index],
					Currency:            drivativeSymbolInfo.Currency[index],
					ExchangeListed:      drivativeSymbolInfo.ExchangeListed[index],
					ExchangeTraded:      drivativeSymbolInfo.ExchangeTraded[index],
					Minmovement:         drivativeSymbolInfo.Minmovement[index],
					Pricescale:          drivativeSymbolInfo.Pricescale[index],
					Timezone:            drivativeSymbolInfo.Timezone[index],
					Type:                drivativeSymbolInfo.Type[index],
					SessionRegular:      drivativeSymbolInfo.SessionRegular[index],
					BaseCurrency:        drivativeSymbolInfo.BaseCurrency[index],
					HasIntraday:         drivativeSymbolInfo.HasIntraday[index],
					Ticker:              drivativeSymbolInfo.Ticker[index],
					IntradayMultipliers: drivativeSymbolInfo.IntradayMultipliers,
					BarFillgaps:         drivativeSymbolInfo.BarFillgaps[index],
				},
			})
			if err != nil {
				cronErrorf("insert derivative symbol info -> symbol:%s: %v", drivativeSymbolInfo.Symbol[index], err)
			}
		}
	}
}

func fetchAndStoreDerivativeSymbols(ctxBg context.Context, svcCtx *svc.ServiceContext, client *injective.Client) {
	defer func() {
		if r := recover(); r != nil {
			cronErrorf("goroutine recovered from fetchAndStoreDerivativeSymbols: %v", r)
		}
	}()
	if release, ok := acquireTaskLock(ctxBg, svcCtx, "derivative_symbols_fetch", 3*time.Second); !ok {
		cronInfof("fetchAndStoreDerivativeSymbols: acquire lock timeout, skip this run")
		return
	} else {
		defer release()
	}
	derivativeSymbolsList, err := getDerivativeSymbolsList(ctxBg, svcCtx, "")
	if err != nil {
		cronErrorf("get derivative symbols list error: %v", err)
		return
	}
	for _, symbol := range derivativeSymbolsList {
		symbols, err := client.DerivativeSymbols(ctxBg, symbol)
		if err != nil {
			cronErrorf("fetch derivative symbols error: %v symbol:%s", err, symbol)
			continue
		}
		filter := bson.M{
			"kind":   "symbols",
			"symbol": symbol,
		}
		count, err := svcCtx.DerivativeColl.CountDocuments(ctxBg, filter)
		if err != nil {
			cronErrorf("count derivative symbols -> symbol:%s: %v", symbol, err)
			continue
		}
		if count == 0 {
			_, err = svcCtx.DerivativeColl.InsertOne(ctxBg, model.DerivativeSymbolsRawDoc{
				Kind:      "symbols",
				Symbol:    symbol,
				Data:      *symbols,
				UpdatedAt: time.Now(),
			})

			if err != nil {
				cronErrorf("insert derivative symbols -> symbol:%s: %v", symbol, err)
			}
		}
	}
}

func getAllDerivativeSymbols(svcCtx *svc.ServiceContext) ([]string, error) {
	filter := bson.M{"kind": "symbols"}
	cur, err := svcCtx.DerivativeColl.Find(context.Background(), filter)
	if err != nil {
		cronErrorf("get all derivative symbols error: %v", err)
		return nil, err
	}
	var symbols []model.DerivativeSymbolsRawDoc
	if err := cur.All(context.Background(), &symbols); err != nil {
		cronErrorf("get all derivative symbols error: %v", err)
		return nil, err
	}
	var symbolList []string
	for _, symbol := range symbols {
		symbolList = append(symbolList, symbol.Symbol)
	}
	return symbolList, nil
}

func fetchAndStoreDerivativeHistory(ctxBg context.Context, svcCtx *svc.ServiceContext, client *injective.Client) {
	defer func() {
		if r := recover(); r != nil {
			cronErrorf("goroutine recovered from fetchAndStoreDerivativeHistory: %v", r)
		}
	}()
	if release, ok := acquireTaskLock(ctxBg, svcCtx, "derivative_history_fetch", 3*time.Second); !ok {
		cronInfof("fetchAndStoreDerivativeHistory: acquire lock timeout, skip this run")
		return
	} else {
		defer release()
	}
	derivativeSymbols, err := getAllDerivativeSymbols(svcCtx)
	if err != nil {
		cronErrorf("fetchAndStoreDerivativeHistory get all derivative symbols error: %v", err)
		return
	}
	if len(derivativeSymbols) == 0 {
		cronErrorf("fetchAndStoreDerivativeHistory get derivative market ids is empty")
		return
	}
	for _, resolution := range append(consts.SupportedMarketResolutions, consts.SupportedDerivativeResolutions...) {
		for _, symbol := range derivativeSymbols {
			var from int64 = 0
			opts := options.FindOne().SetSort(bson.D{{Key: "t", Value: -1}})
			var doc model.DerivativeHistoryRawDoc
			if err := svcCtx.DerivativeColl.FindOne(context.Background(), bson.M{"kind": "history", "symbol": symbol, "resolution": resolution}, opts).Decode(&doc); err == nil {
				from = doc.T
			}
			derivativeHistory, err := client.DerivativeHistory(ctxBg, symbol, resolution, from)
			if err != nil {
				cronErrorf("fetch derivative history error: %v symbol:%s", err, symbol)
				continue
			}
			for index := 0; index < len(derivativeHistory.T); index++ {
				filter := bson.M{
					"kind":       "history",
					"symbol":     symbol,
					"resolution": resolution,
					"t":          derivativeHistory.T[index],
				}
				count, err := svcCtx.DerivativeColl.CountDocuments(ctxBg, filter)
				if err != nil {
					cronErrorf("count derivative history -> symbol:%s resolution:%s t:%d: %v", symbol, resolution, derivativeHistory.T[index], err)
					continue
				}
				if count == 0 {
					_, err = svcCtx.DerivativeColl.InsertOne(ctxBg, model.DerivativeHistoryRawDoc{
						Kind:       "history",
						Symbol:     symbol,
						Resolution: resolution,
						Data: model.DerivativeHistoryRaw{
							C: derivativeHistory.C[index],
							H: derivativeHistory.H[index],
							L: derivativeHistory.L[index],
							O: derivativeHistory.O[index],
							T: derivativeHistory.T[index],
							V: derivativeHistory.V[index],
						},
						T:         derivativeHistory.T[index],
						UpdatedAt: time.Now(),
					})
					if err != nil {
						cronErrorf("insert derivative history -> symbol:%s resolution:%s t:%d: %v", symbol, resolution, derivativeHistory.T[index], err)
					}
				}
			}
		}
	}
}
