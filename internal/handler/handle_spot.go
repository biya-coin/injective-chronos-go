package handler

import (
	"net/http"
	"strconv"

	"github.com/biya-coin/injective-chronos-go/internal/consts"
	"github.com/biya-coin/injective-chronos-go/internal/logic"
	"github.com/biya-coin/injective-chronos-go/internal/model"
	"github.com/biya-coin/injective-chronos-go/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

// SpotConfigHandler proxies Injective spot config with caching via logic layer.
func SpotConfigHandler(ctx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) {
	lgc := logic.NewChartLogic(r.Context(), ctx)
	cfg, err := lgc.GetSpotConfig(r.Context())
	if err != nil {
		logx.Errorf("SpotConfig error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

// SpotMarketHistoryHandler returns candle history for multiple spot marketIDs from Mongo.
// Query: marketIDs=... (repeatable), resolution=5, countback=100
func SpotMarketHistoryHandler(ctx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	marketId := q.Get("marketId")
	resolution := q.Get("resolution")
	if resolution == "" {
		resolution = "1"
	}
	from := q.Get("from")
	var fromInt int64 = 0
	if from != "" && from != "0" {
		fromInt, _ = strconv.ParseInt(from, 10, 64)
	}
	to := q.Get("to")
	if to == "" || to == "0" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing to"})
		return
	}
	var toInt int64 = 0
	toInt, _ = strconv.ParseInt(to, 10, 64)
	// countback optional
	countback := 0
	if v := q.Get("countback"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			countback = n
		}
	}
	if marketId == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing marketId"})
		return
	}
	lgc := logic.NewChartLogic(r.Context(), ctx)
	data, err := lgc.GetMarketHistorySpot(r.Context(), marketId, resolution, countback, fromInt, toInt)
	if err != nil {
		logx.Errorf("SpotMarketHistory error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	// pack response
	writeJSON(w, http.StatusOK, model.SpotMarketHistoryResponse{
		SpotMarketHistory: data,
		S:                 "ok",
	})
}

func SpotMarketSummaryHandler(ctx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) {
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
	resp, err := lgc.GetMarketSummary(r.Context(), consts.MarketTypeSpot, marketId, resolution)
	if err != nil {
		logx.Errorf("SpotMarketSummary error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func SpotMarketSummaryAllHandler(ctx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) {
	lgc := logic.NewChartLogic(r.Context(), ctx)
	resolution := r.URL.Query().Get("resolution")
	if resolution == "" {
		resolution = "24h"
	}
	resp, err := lgc.GetMarketSummaryAll(r.Context(), consts.MarketTypeSpot, resolution)
	if err != nil {
		logx.Errorf("SpotMarketSummaryAll error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func SpotSymbolInfoHandler(ctx *svc.ServiceContext, w http.ResponseWriter, r *http.Request) {
	group := r.URL.Query().Get("group")

	lgc := logic.NewChartLogic(r.Context(), ctx)
	symbolInfo, err := lgc.GetSpotSymbolInfo(r.Context(), group)
	if err != nil {
		logx.Errorf("SpotSymbolInfo error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, symbolInfo)
}
