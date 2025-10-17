package task

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/biya-coin/injective-chronos-go/internal/injective"
	"github.com/biya-coin/injective-chronos-go/internal/svc"
)

func StartCron(ctx *svc.ServiceContext) {
	if !ctx.Config.Cron.Enabled {
		return
	}
	client := injective.NewClient(ctx.Config.Injective, ctx.HttpClient)
	interval := time.Duration(ctx.Config.Cron.IntervalSec) * time.Second
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			<-ticker.C

			logx.Infof("fetching task starting------->spot")

			// Spot
			go fetchAndStoreSpotConfig(context.Background(), ctx, client)
			go fetchAndStoreSpotSummaryAll(context.Background(), ctx, client)
			go fetchAndStoreSpotSummaries(context.Background(), ctx, client)
			go fetchAndStoreSpotMarketHistory(context.Background(), ctx, client)
			go fetchAndStoreSpotSymbolInfo(context.Background(), ctx, client)
			go fetchAndStoreSpotSymbols(context.Background(), ctx, client)

		}
	}()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			<-ticker.C
			protect("cron.tick.derivative", func() {

				// Derivative
				// derivativeResolutions := fetchAndStoreDerivativeConfig(context.Background(), ctx, client)
				logx.Infof("fetching task starting------->derivative")
				fetchAndStoreDerivativeSummaryAll(context.Background(), ctx, client)
				fetchAndStoreDerivativeSummaries(context.Background(), ctx, client)

			})
		}
	}()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			<-ticker.C
			protect("cron.tick.market", func() {

				logx.Infof("fetching task starting------->market")

				// Market
				fetchAndStoreMarketHistory(context.Background(), ctx, client)

			})
		}
	}()
}
