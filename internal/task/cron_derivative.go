package task

import (
	"context"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/biya-coin/injective-chronos-go/internal/consts"
	"github.com/biya-coin/injective-chronos-go/internal/injective"
	"github.com/biya-coin/injective-chronos-go/internal/svc"
)

func fetchAndStoreDerivativeSummaryAll(ctxBg context.Context, svcCtx *svc.ServiceContext, client *injective.Client) {
	for _, res := range consts.SupportedResolutions {
		v, err := client.DerivativeMarketSummaryAll(ctxBg, res)
		if err != nil {
			logx.Errorf("fetch derivative summary_all -> resolution %s: error %v", res, err)
			continue
		}
		_, e := svcCtx.DerivativeColl.InsertOne(ctxBg, bson.M{
			"kind":       "summary_all",
			"resolution": res,
			"data":       v,
			"updated_at": time.Now(),
		})
		if e != nil {
			logx.Errorf("insert derivative summary_all -> resolution %s: error %v", res, e)
		}
	}
}

func fetchAndStoreDerivativeSummaries(ctxBg context.Context, svcCtx *svc.ServiceContext, client *injective.Client) {
	for _, res := range consts.SupportedResolutions {
		marketIds := getMarketSummaryAllIds(svcCtx, res, consts.MarketTypeDerivative)
		if marketIds == nil {
			logx.Errorf("get market derivative summary all ids -> resolution %s: error %v", res, marketIds)
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
					logx.Errorf("fetch derivative summary %s %s: %v", mid, res, err)
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
					logx.Errorf("insert derivative summary %s %s: %v", mid, res, e)
				}
			}(mid)
		}
		wg.Wait()
	}
}
