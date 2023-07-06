package main

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/clintrovert/go-playground/internal/server"
	"github.com/clintrovert/go-playground/pkg/postgres/database"
	"github.com/clintrovert/go-playground/pkg/rediscache"
	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	driver     = "postgres"
	connEnvVar = "POSTGRES_CONN_STR"
	grpcAddr   = ":9099"
	httpAddr   = ":8088"
)

var cacheTtl = time.Hour

func main() {
	metrics := openmetrics.NewRegisteredServerMetrics(
		prometheus.DefaultRegisterer,
		openmetrics.WithServerHandlingTimeHistogram(),
	)
	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics)

	limiter := server.NewRateLimiter()
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(server.Recover),
	}
	cache := getCacheInterceptor()

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
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			openmetrics.UnaryServerInterceptor(metrics),
			auth.UnaryServerInterceptor(server.Authorize),
			ratelimit.UnaryServerInterceptor(limiter),
			recovery.UnaryServerInterceptor(recoveryOpts...),
			cache.UnaryServerInterceptor(rediscache.GenerateRedisKey, cacheTtl),
			server.CustomUnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			openmetrics.StreamServerInterceptor(metrics),
			auth.StreamServerInterceptor(server.Authorize),
			ratelimit.StreamServerInterceptor(limiter),
			recovery.StreamServerInterceptor(recoveryOpts...),
			cache.StreamServerInterceptor(rediscache.GenerateRedisKey, cacheTtl),
			server.CustomStreamInterceptor,
		),
	)
	httpServer := &http.Server{Addr: httpAddr}

	metrics.InitializeMetrics(grpcServer)
	db := getDatabase()

	// Register service RPCs on server
	server.RegisterUserService(grpcServer, db)
	server.RegisterProductService(grpcServer, db)

	// Enable grpc reflection for grpcurl
	reflection.Register(grpcServer)

	g := &run.Group{}
	g.Add(server.ServeGrpc(grpcServer, grpcAddr))
	g.Add(server.ServeHttp(httpServer, registry))

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
