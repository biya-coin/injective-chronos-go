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
			protect("cron.tick", func() {
				logx.Infof("fetching task starting")

				// Spot
				fetchAndStoreSpotConfig(context.Background(), ctx, client)
				fetchAndStoreSpotSummaryAll(context.Background(), ctx, client)
				fetchAndStoreSpotSummaries(context.Background(), ctx, client)
				fetchAndStoreSpotMarketHistory(context.Background(), ctx, client)

				// Derivative
				// derivativeResolutions := fetchAndStoreDerivativeConfig(context.Background(), ctx, client)
				fetchAndStoreDerivativeSummaryAll(context.Background(), ctx, client)
				fetchAndStoreDerivativeSummaries(context.Background(), ctx, client)

				// Market
				fetchAndStoreMarketHistory(context.Background(), ctx, client)

			})
		}
	}()
}
