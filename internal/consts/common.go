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
