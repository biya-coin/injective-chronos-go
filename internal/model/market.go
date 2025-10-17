package model

import "time"

type MarketSummaryCommon struct {
	MarketID string  `json:"marketId" bson:"marketId"`
	Open     float64 `json:"open" bson:"open"`
	High     float64 `json:"high" bson:"high"`
	Low      float64 `json:"low" bson:"low"`
	Volume   float64 `json:"volume" bson:"volume"`
	Price    float64 `json:"price" bson:"price"`
	Change   float64 `json:"change" bson:"change"`
}

type SpotMarketSummary struct {
	MarketSummaryCommon `json:",inline" bson:",inline"`
}

type DerivativeMarketSummary struct {
	MarketSummaryCommon `json:",inline" bson:",inline"`
}

type MarketHistoryRaw struct {
	MarketID   string  `json:"marketID" bson:"marketID"`
	Resolution string  `json:"resolution" bson:"resolution"`
	T          int64   `json:"t" bson:"t"`
	O          float64 `json:"o" bson:"o"`
	H          float64 `json:"h" bson:"h"`
	L          float64 `json:"l" bson:"l"`
	C          float64 `json:"c" bson:"c"`
	V          float64 `json:"v" bson:"v"`
}

type MarketHistoryRawDoc struct {
	Kind       string           `bson:"kind"`
	Resolution string           `bson:"resolution"`
	T          int64            `bson:"t"`
	UpdatedAt  time.Time        `bson:"updated_at"`
	Data       MarketHistoryRaw `bson:"data"`
	MarketId   string           `bson:"marketId"`
}

type MarketHistory struct {
	MarketID   string    `json:"marketID" bson:"marketID"`
	Resolution string    `json:"resolution" bson:"resolution"`
	T          []int64   `json:"t" bson:"t"`
	O          []float64 `json:"o" bson:"o"`
	H          []float64 `json:"h" bson:"h"`
	L          []float64 `json:"l" bson:"l"`
	C          []float64 `json:"c" bson:"c"`
	V          []float64 `json:"v" bson:"v"`
}

// {

// 	"s": "ok"
//   }
