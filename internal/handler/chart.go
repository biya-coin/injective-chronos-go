package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/biya-coin/injective-chronos-go/internal/consts"
	"github.com/biya-coin/injective-chronos-go/internal/logic"
	"github.com/biya-coin/injective-chronos-go/internal/svc"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func DerivativeMarketSummaryAllHandler(ctx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) {
	lgc := logic.NewChartLogic(r.Context(), ctx)
	resolution := r.URL.Query().Get("resolution")
	if resolution == "" {
		resolution = "24h"
	}
	resp, err := lgc.GetMarketSummaryAll(r.Context(), consts.MarketTypeDerivative, resolution)
	if err != nil {
		logx.Errorf("DerivativeMarketSummaryAll error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func DerivativeMarketSummaryHandler(ctx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) {
	marketId := r.URL.Query().Get("marketId")
	resolution := r.URL.Query().Get("resolution")
	if marketId == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing marketId query param"})
		return
	}
	if resolution == "" {
		resolution = "24h"
	}
	lgc := logic.NewChartLogic(r.Context(), ctx)
	resp, err := lgc.GetMarketSummary(r.Context(), consts.MarketTypeDerivative, marketId, resolution)
	if err != nil {
		logx.Errorf("DerivativeMarketSummary error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// MarketHistoryHandler returns candle history for multiple marketIDs from Mongo.
// Query: marketIDs=... (repeatable), resolution=5, countback=100
func MarketHistoryHandler(ctx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	marketIDs := q["marketIDs"]
	resolution := q.Get("resolution")
	if resolution == "" {
		resolution = "5"
	}
	// countback optional
	countback := 0
	if v := q.Get("countback"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			countback = n
		}
	}

	if len(marketIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing marketIDs"})
		return
	}
	lgc := logic.NewChartLogic(r.Context(), ctx)
	data, err := lgc.GetMarketHistory(r.Context(), marketIDs, resolution, countback)
	if err != nil {
		logx.Errorf("MarketHistory error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, data)
}

// DerivativeMarketHistoryHandler returns candle history for multiple derivative marketIDs from Mongo.
// Query: marketIDs=... (repeatable), resolution=5, countback=100
func DerivativeMarketHistoryHandler(ctx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	symbol := q.Get("symbol")
	marketId := q.Get("marketId")
	resolution := q.Get("resolution")
	if resolution == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing resolution query param"})
		return
	}
	// countback optional
	countback := 0
	if v := q.Get("countback"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			countback = n
		}
	}

	from := q.Get("from")
	var fromInt int64 = 0
	if from != "" && from != "0" {
		fromInt, _ = strconv.ParseInt(from, 10, 64)
	}
	to := q.Get("to")
	if to == "" || to == "0" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing to  query param"})
		return
	}
	var toInt int64 = 0
	toInt, _ = strconv.ParseInt(to, 10, 64)
	if marketId == "" && symbol == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing marketId or symbol query param"})
		return
	}
	lgc := logic.NewChartLogic(r.Context(), ctx)
	data, err := lgc.GetMarketHistoryDerivative(r.Context(), symbol, marketId, resolution, countback, fromInt, toInt)
	if err != nil {
		logx.Errorf("GetMarketHistoryDerivative error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, model.DerivativeMarketHistoryResponse{
		DerivativeMarketHistory: data,
		S:                       "ok",
	})
}

// DerivativeConfigHandler proxies Injective derivative config with caching via logic layer.
func DerivativeConfigHandler(ctx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) {
	lgc := logic.NewChartLogic(r.Context(), ctx)
	cfg, err := lgc.GetDerivativeConfig(r.Context())
	if err != nil {
		logx.Errorf("DerivativeConfig error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func DerivativeSymbolInfoHandler(ctx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) {
	lgc := logic.NewChartLogic(r.Context(), ctx)
	group := r.URL.Query().Get("group")
	symbolInfo, err := lgc.GetDerivativeSymbolInfo(r.Context(), group)
	if err != nil {
		logx.Errorf("DerivativeSymbolInfo error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, symbolInfo)
}

func DerivativeSymbolsHandler(ctx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) {
	lgc := logic.NewChartLogic(r.Context(), ctx)
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing symbol query param"})
		return
	}
	symbols, err := lgc.GetDerivativeSymbols(r.Context(), symbol)
	if err != nil {
		logx.Errorf("DerivativeSymbols error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, symbols)
}
