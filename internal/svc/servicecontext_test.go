package svc

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/biya-coin/injective-chronos-go/internal/config"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

// This test verifies connectivity to Redis and Mongo using env-provided settings.
// It is skipped when required env vars are not set.
func TestNewServiceContext_Connectivity(t *testing.T) {
	var c config.Config
	configFile := getenv("TEST_CONFIG_FILE", "../../etc/config.yaml")
	conf.MustLoad(configFile, &c)
	logx.Infof("loaded config: %+v", c)
	s := NewServiceContext(c)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Redis ping
	if err := s.Redis.Ping(ctx).Err(); err != nil {
		t.Fatalf("redis ping failed: %v", err)
	}
	// Mongo ping and basic collection check by counting 0 docs (should not error)
	if err := s.MongoClient.Ping(ctx, nil); err != nil {
		t.Fatalf("mongo ping failed: %v", err)
	}
	if _, err := s.SpotColl.CountDocuments(ctx, struct{}{}); err != nil {
		t.Fatalf("spot collection count failed: %v", err)
	}
	if _, err := s.DerivativeColl.CountDocuments(ctx, struct{}{}); err != nil {
		t.Fatalf("derivative collection count failed: %v", err)
	}
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
