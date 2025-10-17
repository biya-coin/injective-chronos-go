package config

import (
	"github.com/zeromicro/go-zero/rest"
)

type RedisConf struct {
	Address        string
	Password       string
	DB             int
	TTLSeconds     int // base TTL for caches
	JitterSeconds  int // random TTL jitter to avoid herd expiry
	LockTTLSeconds int // distributed lock TTL
	RetryMs        int // poll interval when waiting others to fill cache
	RetryMax       int // max polls
}

type MongoCollections struct {
	Spot       string
	Derivative string
	Market     string
}

type MongoConf struct {
	URI         string
	Database    string
	Collections MongoCollections
}

type InjectiveConf struct {
	BaseURL   string
	TimeoutMs int
}

type CronConf struct {
	Enabled     bool
	IntervalSec int
}

type Config struct {
	rest.RestConf
	Redis     RedisConf
	Mongo     MongoConf
	Injective InjectiveConf
	Cron      CronConf
}
