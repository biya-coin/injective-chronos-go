package task

import (
	"context"
	"time"

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

			cronInfof("fetching task starting------->spot")

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
				cronInfof("fetching task starting------->derivative")
				go fetchAndStoreDerivativeSummaryAll(context.Background(), ctx, client)
				go fetchAndStoreDerivativeSummaries(context.Background(), ctx, client)
				go fetchAndStoreDerivativeSymbolInfo(context.Background(), ctx, client)
				go fetchAndStoreDerivativeSymbols(context.Background(), ctx, client)

			})
		}
	}()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			<-ticker.C
			protect("cron.tick.market", func() {

				cronInfof("fetching task starting------->market")

				// Market
				fetchAndStoreMarketHistory(context.Background(), ctx, client)

			})
		}
	}()
}
