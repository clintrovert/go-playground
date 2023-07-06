package main

import (
	"database/sql"
	"os"

	"github.com/clintrovert/go-playground/internal/server"
	"github.com/clintrovert/go-playground/pkg/grpc"
	"github.com/clintrovert/go-playground/pkg/postgres/database"
	"github.com/clintrovert/go-playground/pkg/rediscache"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/oklog/run"
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

	limiter := server.NewRateLimiter()
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(server.Recover),
	}
	// cache := getCacheInterceptor()

	srv := grpc.NewServer(grpcAddr, httpAddr).
		WithAuth(server.Authorize).
		WithRecovery(recoveryOpts).
		WithRateLimiter(limiter).
		WithReflection().
		WithDefaultMetrics()

	db := getDatabase()

	// Register service RPCs on server
	server.RegisterUserService(srv.GrpcServer(), db)
	server.RegisterProductService(srv.GrpcServer(), db)

	g := &run.Group{}
	g.Add(srv.ServeGrpc())
	g.Add(srv.ServeHttp())

	if err := g.Run(); err != nil {
		os.Exit(1)
	}
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

func getCacheInterceptor() *server.CacheInterceptor {
	rdb := rediscache.NewRedisCache()
	kvc := server.NewKeyValCacheInterceptor(rdb, logrus.NewEntry(logrus.New()))
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
//grpcServer := grpc.NewServer(
//	grpc.ChainUnaryInterceptor(
//		openmetrics.UnaryServerInterceptor(metrics),
//		auth.UnaryServerInterceptor(server.Authorize),
//		ratelimit.UnaryServerInterceptor(limiter),
//		recovery.UnaryServerInterceptor(recoveryOpts...),
//		cache.UnaryServerInterceptor(rediscache.GenerateRedisKey, cacheTtl),
//		server.CustomUnaryInterceptor,
//	),
//	grpc.ChainStreamInterceptor(
//		openmetrics.StreamServerInterceptor(metrics),
//		auth.StreamServerInterceptor(server.Authorize),
//		ratelimit.StreamServerInterceptor(limiter),
//		recovery.StreamServerInterceptor(recoveryOpts...),
//		cache.StreamServerInterceptor(rediscache.GenerateRedisKey, cacheTtl),
//		server.CustomStreamInterceptor,
//	),
//)
