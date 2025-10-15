package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/biya-coin/injective-chronos-go/internal/consts"
	"github.com/biya-coin/injective-chronos-go/internal/svc"
)

type ChartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChartLogic {
	return &ChartLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *ChartLogic) GetMarketSummaryAll(ctx context.Context, marketType consts.MarketType, resolution string) (interface{}, error) {

	if marketType == consts.MarketTypeDerivative {
		return l.getMarketSummaryAllDerivative(ctx, resolution)
	}
	if marketType == consts.MarketTypeSpot {
		return l.getMarketSummaryAllSpot(ctx, resolution)
	}
	return nil, errors.New("invalid market type")
}

func (l *ChartLogic) GetMarketSummary(ctx context.Context, marketType consts.MarketType, marketId string, resolution string) (interface{}, error) {
	if marketId == "" {
		return nil, errors.New("empty marketId")
	}
	if marketType == consts.MarketTypeDerivative {
		return l.getMarketSummaryDerivative(ctx, marketId, resolution)
	}
	if marketType == consts.MarketTypeSpot {
		return l.getMarketSummarySpot(ctx, marketId, resolution)
	}
	return nil, errors.New("invalid market type")
}
