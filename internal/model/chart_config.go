package model

// ChartConfig represents the TradingView-style chart configuration returned by Injective.
// It is used for both spot and derivative config responses.
type ChartSpotConfig struct {
	SupportedResolutions   []string `json:"supported_resolutions"`
	SupportsGroupRequest   bool     `json:"supports_group_request"`
	SupportsMarks          bool     `json:"supports_marks"`
	SupportsSearch         bool     `json:"supports_search"`
	SupportsTimescaleMarks bool     `json:"supports_timescale_marks"`
}

type ChartDerivativeConfig struct {
	SupportedResolutions   []string `json:"supported_resolutions"`
	SupportsGroupRequest   bool     `json:"supports_group_request"`
	SupportsMarks          bool     `json:"supports_marks"`
	SupportsSearch         bool     `json:"supports_search"`
	SupportsTimescaleMarks bool     `json:"supports_timescale_marks"`
}
