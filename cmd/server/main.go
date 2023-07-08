package main

import (
	"database/sql"
	"os"

	"github.com/clintrovert/go-playground/internal/playground"
	"github.com/clintrovert/go-playground/pkg/postgres/database"
	"github.com/clintrovert/go-playground/pkg/rediscache"
	"github.com/clintrovert/go-playground/pkg/server"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/sirupsen/logrus"
)

const (
	driver     = "postgres"
	connEnvVar = "POSTGRES_CONN_STR"
	grpcAddr   = ":9099"
	httpAddr   = ":8088"
)

//var cacheTtl = time.Hour

func main() {

	limiter := playground.NewRateLimiter()
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(playground.Recover),
	}
	// cache := getCacheInterceptor()

	srv := server.NewBuilder(grpcAddr, httpAddr).
		WithDefaultMetrics().
		WithAuth(playground.Authorize).
		WithRecovery(recoveryOpts).
		WithRateLimiter(limiter).
		WithGrpcReflection().
		Build()

	db := getDatabase()

	// Register service RPCs on playground
	playground.RegisterUserService(srv.Grpc, db)
	playground.RegisterProductService(srv.Grpc, db)

	srv.Serve()
}

func getDatabase() *database.Queries {
	// Open a connection to the database cluster.
	connStr := os.Getenv(connEnvVar)
	postgres, err := sql.Open(driver, connStr)
	if err != nil {
		panic(err)
	}

	return database.New(postgres)
}

func getCacheInterceptor() *playground.CacheInterceptor {
	rdb := rediscache.NewRedisCache()
	kvc := playground.NewKeyValCacheInterceptor(
		rdb, logrus.NewEntry(logrus.New()),
	)
	return kvc
}

// Set up the following middlewares on unary/stream requests, (ordering of
// these matters to some extent):
// - metrics
// - auth
// - rate limiting
// - logging
// - req validation
// - tracing
// - caching
// - custom
//grpcServer := grpc.NewBuilder(
//	grpc.ChainUnaryInterceptor(
//		openmetrics.UnaryServerInterceptor(metrics),
//		auth.UnaryServerInterceptor(playground.Authorize),
//		ratelimit.UnaryServerInterceptor(limiter),
//		recovery.UnaryServerInterceptor(recoveryOpts...),
//		cache.UnaryServerInterceptor(rediscache.GenerateRedisKey, cacheTtl),
//		playground.CustomUnaryInterceptor,
//	),
//	grpc.ChainStreamInterceptor(
//		openmetrics.StreamServerInterceptor(metrics),
//		auth.StreamServerInterceptor(playground.Authorize),
//		ratelimit.StreamServerInterceptor(limiter),
//		recovery.StreamServerInterceptor(recoveryOpts...),
//		cache.StreamServerInterceptor(rediscache.GenerateRedisKey, cacheTtl),
//		playground.CustomStreamInterceptor,
//	),
//)
