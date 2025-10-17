package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/biya-coin/injective-chronos-go/internal/logic"
	"github.com/biya-coin/injective-chronos-go/internal/svc"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
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
