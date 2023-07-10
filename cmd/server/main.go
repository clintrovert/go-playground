package main

import (
	"database/sql"
	"os"
	"time"

	"github.com/clintrovert/go-playground/internal/playground"
	"github.com/clintrovert/go-playground/pkg/postgres/database"
	"github.com/clintrovert/go-playground/pkg/redis"
	"github.com/clintrovert/go-playground/pkg/server"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	driver     = "postgres"
	connEnvVar = "POSTGRES_CONN_STR"
	grpcAddr   = ":9099"
	httpAddr   = ":8088"
)

var cacheTtl = time.Hour

func main() {
	limiter := playground.NewRateLimiter()
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(playground.Recover),
	}
	rdb := redis.NewRedisCache()

	srv, err := server.NewBuilder(grpcAddr, httpAddr).
		WithMetrics(prometheus.DefaultRegisterer).
		WithCache(rdb, redis.GenerateKeyFromRpc, cacheTtl).
		WithAuth(playground.Authorize).
		WithRecovery(recoveryOpts).
		WithRateLimiter(limiter).
		WithGrpcReflection().
		WithGrpcValidation().
		Build()

	if err != nil {
		panic(err)
	}

	db := getDatabase()

	// Register service RPCs on playground
	playground.RegisterUserService(srv.GrpcServer, db)
	playground.RegisterProductService(srv.GrpcServer, db)

	srv.HttpServer.ReadHeaderTimeout = time.Second * 2

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
//		cache.UnaryServerInterceptor(redis.GenerateRedisKey, cacheTtl),
//		playground.CustomUnaryInterceptor,
//	),
//	grpc.ChainStreamInterceptor(
//		openmetrics.StreamServerInterceptor(metrics),
//		auth.StreamServerInterceptor(playground.Authorize),
//		ratelimit.StreamServerInterceptor(limiter),
//		recovery.StreamServerInterceptor(recoveryOpts...),
//		cache.StreamServerInterceptor(redis.GenerateRedisKey, cacheTtl),
//		playground.CustomStreamInterceptor,
//	),
//)
