package svc

import (
	"context"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/biya-coin/injective-chronos-go/internal/config"
	"github.com/biya-coin/injective-chronos-go/internal/logutil"
)

type ServiceContext struct {
	Config         config.Config
	Redis          *redis.Client
	MongoClient    *mongo.Client
	SpotColl       *mongo.Collection
	DerivativeColl *mongo.Collection
	MarketColl     *mongo.Collection
	HttpClient     *http.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Address,
		Password: c.Redis.Password,
		DB:       c.Redis.DB,
	})

	// Mongo
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(c.Mongo.URI))
	if err != nil {
		logx.Errorf("failed to connect mongo: %v", err)
		panic(err)
	}

	// Ensure database exists; create by creating required collections if missing
	dbName := c.Mongo.Database
	dbNames, dErr := client.ListDatabaseNames(context.Background(), bson.M{})
	if dErr != nil {
		logx.Errorf("list mongo databases failed: %v", dErr)
	}
	existsDB := false
	for _, n := range dbNames {
		if n == dbName {
			existsDB = true
			break
		}
	}
	if !existsDB {
		// Create DB implicitly by creating necessary collections
		_ = client.Database(dbName).CreateCollection(context.Background(), c.Mongo.Collections.Spot)
		_ = client.Database(dbName).CreateCollection(context.Background(), c.Mongo.Collections.Derivative)
		if c.Mongo.Collections.Market != "" {
			_ = client.Database(dbName).CreateCollection(context.Background(), c.Mongo.Collections.Market)
		}
	}

	// Ensure collections exist (create if missing)
	db := client.Database(c.Mongo.Database)
	collNames, lerr := db.ListCollectionNames(context.Background(), bson.M{})
	if lerr != nil {
		logx.Errorf("list mongo collections failed: %v", lerr)
	}
	existsSpot := false
	existsDerivative := false
	existsMarket := false
	for _, n := range collNames {
		if n == c.Mongo.Collections.Spot {
			existsSpot = true
		}
		if n == c.Mongo.Collections.Derivative {
			existsDerivative = true
		}
		if n == c.Mongo.Collections.Market {
			existsMarket = true
		}
	}
	if !existsSpot {
		if e := db.CreateCollection(context.Background(), c.Mongo.Collections.Spot); e != nil {
			logx.Errorf("create collection %s failed: %v", c.Mongo.Collections.Spot, e)
		}
	}
	if !existsDerivative {
		if e := db.CreateCollection(context.Background(), c.Mongo.Collections.Derivative); e != nil {
			logx.Errorf("create collection %s failed: %v", c.Mongo.Collections.Derivative, e)
		}
	}
	if !existsMarket && c.Mongo.Collections.Market != "" {
		if e := db.CreateCollection(context.Background(), c.Mongo.Collections.Market); e != nil {
			logx.Errorf("create collection %s failed: %v", c.Mongo.Collections.Market, e)
		}
	}

	spot := db.Collection(c.Mongo.Collections.Spot)
	derivative := db.Collection(c.Mongo.Collections.Derivative)
	var market *mongo.Collection
	if c.Mongo.Collections.Market != "" {
		market = db.Collection(c.Mongo.Collections.Market)
	}

	// HTTP client
	hc := &http.Client{Timeout: time.Duration(c.Injective.TimeoutMs) * time.Millisecond}

	// Setup split log writer: api.log for API logs, cron.log for cron logs
	if sw, err := logutil.NewSplitWriter("logs/api.log", "logs/cron.log"); err != nil {
		logx.Errorf("init split log writer failed: %v", err)
	} else {
		logx.SetWriter(logx.NewWriter(sw))
	}

	return &ServiceContext{
		Config:         c,
		Redis:          rdb,
		MongoClient:    client,
		SpotColl:       spot,
		DerivativeColl: derivative,
		MarketColl:     market,
		HttpClient:     hc,
	}
}
