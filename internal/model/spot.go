package model

import (
	"time"
)

type SpotMarketHistoryRaw struct {
	T int64   `json:"t" bson:"t"`
	O float64 `json:"o" bson:"o"`
	H float64 `json:"h" bson:"h"`
	L float64 `json:"l" bson:"l"`
	C float64 `json:"c" bson:"c"`
	V float64 `json:"v" bson:"v"`
}

type SpotMarketHistory struct {
	T []int64   `json:"t" bson:"t"`
	O []float64 `json:"o" bson:"o"`
	H []float64 `json:"h" bson:"h"`
	L []float64 `json:"l" bson:"l"`
	C []float64 `json:"c" bson:"c"`
	V []float64 `json:"v" bson:"v"`
}

type SpotMarketHistoryResponse struct {
	SpotMarketHistory `json:",inline" bson:",inline"`
	S                 string `json:"s" bson:"s"`
}

type SpotHistoryDoc struct {
	Kind       string               `bson:"kind"`
	Resolution string               `bson:"resolution"`
	T          int64                `bson:"t"`
	UpdatedAt  time.Time            `bson:"updated_at"`
	Data       SpotMarketHistoryRaw `bson:"data"`
}
