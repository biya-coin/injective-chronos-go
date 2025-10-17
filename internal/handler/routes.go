package handler

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/rest"

	"github.com/biya-coin/injective-chronos-go/internal/consts"
	"github.com/biya-coin/injective-chronos-go/internal/svc"
)

// health check
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"ts": time.Now().Unix()})
}

func RegisterHandlers(server *rest.Server, ctx *svc.ServiceContext) {
	// 全局设置 CORS 响应头
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next(w, r)
		}
	})

	// spot
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/api/chart/v1/spot/market_summary_all",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			SpotMarketSummaryAllHandler(ctx, w, r)
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/api/chart/v1/spot/config",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			SpotConfigHandler(ctx, w, r)
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/api/chart/v1/spot/market_summary",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			SpotMarketSummaryHandler(ctx, w, r)
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/api/chart/v1/spot/history",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			SpotMarketHistoryHandler(ctx, w, r)
		},
	})

	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   consts.SpotSymbolInfoPath,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			SpotSymbolInfoHandler(ctx, w, r)
		},
	})

	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   consts.SpotSymbolsPath,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			SpotSymbolsHandler(ctx, w, r)
		},
	})
	// derivative
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/api/chart/v1/derivative/market_summary_all",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			DerivativeMarketSummaryAllHandler(ctx, w, r)
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/api/chart/v1/derivative/market_summary",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			DerivativeMarketSummaryHandler(ctx, w, r)
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   consts.DerivativeConfigPath,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			DerivativeConfigHandler(ctx, w, r)
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   consts.DerivativeHistoryPath,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			DerivativeMarketHistoryHandler(ctx, w, r)
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   consts.DerivativeSymbolInfoPath,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			DerivativeSymbolInfoHandler(ctx, w, r)
		},
	})
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   consts.DerivativeSymbolsPath,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			DerivativeSymbolsHandler(ctx, w, r)
		},
	})

	// market
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/api/chart/v1/market/history",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			MarketHistoryHandler(ctx, w, r)
		},
	})
}
