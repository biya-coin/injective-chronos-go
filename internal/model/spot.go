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

type SpotSymbolInfoResponse struct {
	SpotSymbolInfo `json:",inline" bson:",inline"`
	S              string `json:"s" bson:"s"`
}

type SpotSymbolsRaw struct {
	Symbol               string   `json:"symbol" bson:"symbol"`
	Ticker               string   `json:"ticker" bson:"ticker"`
	Name                 string   `json:"name" bson:"name"`
	Description          string   `json:"description" bson:"description"`
	Type                 string   `json:"type" bson:"type"`
	Session              string   `json:"session" bson:"session"`
	Minmov               int      `json:"minmov" bson:"minmov"`
	Minmov2              int      `json:"minmov2" bson:"minmov2"`
	Pricescale           int      `json:"pricescale" bson:"pricescale"`
	Fractional           bool     `json:"fractional" bson:"fractional"`
	HasIntraday          bool     `json:"has_intraday" bson:"has_intraday"`
	SupportedResolutions []string `json:"supported_resolutions" bson:"supported_resolutions"`
	IntradayMultipliers  []string `json:"intraday_multipliers" bson:"intraday_multipliers"`
	HasSeconds           bool     `json:"has_seconds" bson:"has_seconds"`
	SecondsMultipliers   []string `json:"seconds_multipliers" bson:"seconds_multipliers"`
	HasDaily             bool     `json:"has_daily" bson:"has_daily"`
	HasWeeklyAndMonthly  bool     `json:"has_weekly_and_monthly" bson:"has_weekly_and_monthly"`
	HasEmptyBars         bool     `json:"has_empty_bars" bson:"has_empty_bars"`
	ForceSessionRebuild  bool     `json:"force_session_rebuild" bson:"force_session_rebuild"`
	HasNoVolume          bool     `json:"has_no_volume" bson:"has_no_volume"`
	VolumePrecision      int      `json:"volume_precision" bson:"volume_precision"`
	DataStatus           string   `json:"data_status" bson:"data_status"`
	Expired              bool     `json:"expired" bson:"expired"`
	CurrencyCode         string   `json:"currency_code" bson:"currency_code"`
}

type SpotSymbolsRawDoc struct {
	Kind      string         `bson:"kind"`
	Symbol    string         `bson:"symbol"`
	UpdatedAt time.Time      `bson:"updated_at"`
	Data      SpotSymbolsRaw `bson:"data"`
}
