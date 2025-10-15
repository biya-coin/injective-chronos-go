package task

import (
	"context"
	"encoding/json"
	"runtime/debug"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/biya-coin/injective-chronos-go/internal/consts"
	"github.com/biya-coin/injective-chronos-go/internal/model"
	"github.com/biya-coin/injective-chronos-go/internal/svc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func parseMarketSummaryAllIds(v []model.MarketSummaryCommon) []string {
	// logx.Infof("parse market summary all ids: %v", v)
	var ids []string
	for _, row := range v {
		if row.MarketID != "" {
			ids = append(ids, row.MarketID)
		}
	}
	return ids
}

func getMarketSummaryAllIds(svcCtx *svc.ServiceContext, resolution string, marketType consts.MarketType) []string {
	// 这里通过mongoch获取resolution对应的marketIds
	opts := options.FindOne().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	var doc bson.M
	if marketType == consts.MarketTypeSpot {
		if err := svcCtx.SpotColl.FindOne(context.Background(), bson.M{"kind": "summary_all", "resolution": resolution}, opts).Decode(&doc); err != nil {
			logx.Errorf("get market spot summary all ids -> resolution %s: error %v", resolution, err)
			return nil
		}
	}
	if marketType == consts.MarketTypeDerivative {
		if err := svcCtx.DerivativeColl.FindOne(context.Background(), bson.M{"kind": "summary_all", "resolution": resolution}, opts).Decode(&doc); err != nil {
			logx.Errorf("get market derivative summary all ids -> resolution %s: error %v", resolution, err)
			return nil
		}
	}
	if doc == nil {
		logx.Errorf("get market summary all ids -> resolution %s: no data", resolution)
		return nil
	}
	bytes, _ := json.Marshal(doc["data"])
	var v = []model.MarketSummaryCommon{}
	err := json.Unmarshal(bytes, &v)
	if err != nil {
		logx.Errorf("get market summary all ids -> resolution %s: error %v", resolution, err)
		return nil
	}
	return parseMarketSummaryAllIds(v)
}

// recoverAndLog recovers from panic and logs error with stack trace and caller info.
func recoverAndLog(where string) {
	if r := recover(); r != nil {
		logx.Errorf("panic recovered at %s: %v\nstack:\n%s", where, r, string(debug.Stack()))
	}
}

// protect wraps a function call with panic recovery; where describes the context.
func protect(where string, fn func()) {
	defer recoverAndLog(where)
	fn()
}
