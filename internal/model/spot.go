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

type SpotSymbolInfoRaw struct {
	Symbol              string   `json:"symbol" bson:"symbol"`
	Name                string   `json:"name" bson:"name"`
	Description         string   `json:"description" bson:"description"`
	Currency            string   `json:"currency" bson:"currency"`
	ExchangeListed      string   `json:"exchange-listed" bson:"exchange-listed"`
	ExchangeTraded      string   `json:"exchange-traded" bson:"exchange-traded"`
	Minmovement         int      `json:"minmovement" bson:"minmovement"`
	Pricescale          int      `json:"pricescale" bson:"pricescale"`
	Timezone            string   `json:"timezone" bson:"timezone"`
	Type                string   `json:"type" bson:"type"`
	SessionRegular      string   `json:"session-regular" bson:"session-regular"`
	BaseCurrency        string   `json:"base-currency" bson:"base-currency"`
	HasIntraday         bool     `json:"has-intraday" bson:"has-intraday"`
	Ticker              string   `json:"ticker" bson:"ticker"`
	IntradayMultipliers []string `json:"intraday-multipliers" bson:"intraday-multipliers"`
	BarFillgaps         bool     `json:"bar-fillgaps" bson:"bar-fillgaps"`
}
type SpotSymbolInfo struct {
	Symbol              []string `json:"symbol" bson:"symbol"`
	Name                []string `json:"name" bson:"name"`
	Description         []string `json:"description" bson:"description"`
	Currency            []string `json:"currency" bson:"currency"`
	ExchangeListed      []string `json:"exchange-listed" bson:"exchange-listed"`
	ExchangeTraded      []string `json:"exchange-traded" bson:"exchange-traded"`
	Minmovement         []int    `json:"minmovement" bson:"minmovement"`
	Pricescale          []int    `json:"pricescale" bson:"pricescale"`
	Timezone            []string `json:"timezone" bson:"timezone"`
	Type                []string `json:"type" bson:"type"`
	SessionRegular      []string `json:"session-regular" bson:"session-regular"`
	BaseCurrency        []string `json:"base-currency" bson:"base-currency"`
	HasIntraday         []bool   `json:"has-intraday" bson:"has-intraday"`
	Ticker              []string `json:"ticker" bson:"ticker"`
	IntradayMultipliers []string `json:"intraday-multipliers" bson:"intraday-multipliers"`
	BarFillgaps         []bool   `json:"bar-fillgaps" bson:"bar-fillgaps"`
}

type SpotSymbolInfoRawDoc struct {
	Kind      string            `bson:"kind"`
	Symbol    string            `bson:"symbol"`
	Group     string            `bson:"group"`
	UpdatedAt time.Time         `bson:"updated_at"`
	Data      SpotSymbolInfoRaw `bson:"data"`
}
