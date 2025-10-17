package cache

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := crand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func acquireLock(ctx context.Context, rdb *redis.Client, lockKey, token string, ttl time.Duration) bool {
	ok, err := rdb.SetNX(ctx, lockKey, token, ttl).Result()
	if err != nil {
		logx.Errorf("cache.acquireLock error: %v", err)
		return false
	}
	return ok
}

var releaseLockScript = redis.NewScript(`
if redis.call('get', KEYS[1]) == ARGV[1] then
    return redis.call('del', KEYS[1])
else
    return 0
end`)

func releaseLock(ctx context.Context, rdb *redis.Client, lockKey, token string) {
	// 使用独立的短超时后台上下文，避免受请求上下文取消/超时影响
	relCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if _, err := releaseLockScript.Run(relCtx, rdb, []string{lockKey}, token).Result(); err != nil {
		logx.Errorf("cache.releaseLock error: %v", err)
	}
}

// GetOrLoadBytes returns cached bytes, or uses loader under a distributed lock to prevent stampede.
func GetOrLoadBytes(
	ctx context.Context,
	rdb *redis.Client,
	key string,
	baseTTLSeconds int,
	jitterSeconds int,
	lockTTLSeconds int,
	retryMs int,
	retryMax int,
	loader func(context.Context) ([]byte, error),
) ([]byte, error) {
	if s, err := rdb.Get(ctx, key).Bytes(); err == nil && len(s) > 0 {
		return s, nil
	}

	lockKey := "lock:" + key
	token, terr := generateToken()
	if terr != nil {
		logx.Errorf("cache.generateToken error: %v", terr)
		// continue without token but avoid deleting others' locks; we will not acquire
	}

	gotLock := false
	if token != "" {
		gotLock = acquireLock(ctx, rdb, lockKey, token, time.Duration(lockTTLSeconds)*time.Second)
	}

	if gotLock {
		defer releaseLock(ctx, rdb, lockKey, token)
		bytes, err := loader(ctx)
		if err != nil {
			return nil, err
		}
		ttl := baseTTLSeconds
		if jitterSeconds > 0 {
			ttl += rand.Intn(jitterSeconds + 1)
		}
		_ = rdb.Set(ctx, key, bytes, time.Duration(ttl)*time.Second).Err()
		return bytes, nil
	}

	// Wait for the holder to populate cache
	backoff := time.Duration(retryMs) * time.Millisecond
	for i := 0; i < retryMax; i++ {
		time.Sleep(backoff)
		if s, err := rdb.Get(ctx, key).Bytes(); err == nil && len(s) > 0 {
			return s, nil
		}
	}
	// still miss; give up to avoid DB thundering herd
	return nil, nil
}
