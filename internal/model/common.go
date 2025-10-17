package model

type MarketSummaryCommon struct {
	MarketID string  `json:"marketId" bson:"marketId"`
	Open     float64 `json:"open" bson:"open"`
	High     float64 `json:"high" bson:"high"`
	Low      float64 `json:"low" bson:"low"`
	Volume   float64 `json:"volume" bson:"volume"`
	Price    float64 `json:"price" bson:"price"`
	Change   float64 `json:"change" bson:"change"`
}
