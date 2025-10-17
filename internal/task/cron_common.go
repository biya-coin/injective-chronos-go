package task

import (
	"context"
	"encoding/json"
	"runtime/debug"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/biya-coin/injective-chronos-go/internal/consts"
	"github.com/biya-coin/injective-chronos-go/internal/model"
	"github.com/biya-coin/injective-chronos-go/internal/svc"
)

// acquireTaskLock 尝试在 wait 时间内基于 Redis 获取分布式任务锁。
// 成功返回释放函数与 true；失败返回 nil、false。
func acquireTaskLock(ctx context.Context, svcCtx *svc.ServiceContext, key string, wait time.Duration) (func(), bool) {
	if key == "" {
		return nil, false
	}
	lockKey := "lock:task:" + key

	// TTL 与重试间隔从配置读取，提供合理默认
	lockTTL := time.Duration(svcCtx.Config.Redis.LockTTLSeconds) * time.Second
	if lockTTL <= 0 {
		lockTTL = 60 * time.Second
	}
	retryMs := svcCtx.Config.Redis.RetryMs
	if retryMs <= 0 {
		retryMs = 100
	}
	retryInterval := time.Duration(retryMs) * time.Millisecond

	// 获取带超时的上下文
	lockCtx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	for {
		select {
		case <-lockCtx.Done():
			logx.Infof("acquireTaskLock timeout for key=%s", key)
			return nil, false
		default:
		}

		ok, err := svcCtx.Redis.SetNX(lockCtx, lockKey, "1", lockTTL).Result()
		if err != nil {
			logx.Errorf("acquireTaskLock SetNX error for key=%s: %v", key, err)
			return nil, false
		}
		if ok {
			// 成功，提供释放函数
			release := func() {
				if _, err := svcCtx.Redis.Del(context.Background(), lockKey).Result(); err != nil {
					logx.Errorf("releaseTaskLock DEL error for key=%s: %v", key, err)
				}
			}
			return release, true
		}
		time.Sleep(retryInterval)
	}
}

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

// cron logs helpers to add [CRON] prefix for split writer routing
func cronInfof(format string, args ...any) {
	logx.Infof("[CRON] "+format, args...)
}

func cronErrorf(format string, args ...any) {
	logx.Errorf("[CRON] "+format, args...)
}
