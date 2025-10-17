package consts

var (
	SupportedResolutions       = []string{"hour", "60m", "day", "24h", "week", "7days", "month", "30days"}
	SupportedMarketResolutions = []string{"1", "5", "15", "30", "60", "120", "240", "720", "1440"}
)

type MarketType string

const (
	MarketTypeSpot       MarketType = "spot"
	MarketTypeDerivative MarketType = "derivative"
)

const (

	// spot
	SpotSummaryAllPath = "/api/chart/v1/spot/market_summary_all"
	SpotSummaryPath    = "/api/chart/v1/spot/market_summary"
	SpotConfigPath     = "/api/chart/v1/spot/config"
	SpotHistoryPath    = "/api/chart/v1/spot/history"
	SpotSymbolInfoPath = "/api/chart/v1/spot/symbol_info"
	SpotSymbolsPath    = "/api/chart/v1/spot/symbols"

	// derivative
	DerivativeSummaryAllPath = "/api/chart/v1/derivative/market_summary_all"
	DerivativeSummaryPath    = "/api/chart/v1/derivative/market_summary"
	DerivativeConfigPath     = "/api/chart/v1/derivative/config"
	DerivativeSymbolInfoPath = "/api/chart/v1/derivative/symbol_info"
	DerivativeSymbolsPath    = "/api/chart/v1/derivative/symbols"
	DerivativeHistoryPath    = "/api/chart/v1/derivative/history"
	// market
	MarketHistoryPath = "/api/chart/v1/market/history"
)
