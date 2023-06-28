package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"

	"github.com/clintrovert/go-playground/internal/postgres/database"
	"github.com/clintrovert/go-playground/internal/server"
	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	driver     = "postgres"
	connEnvVar = "POSTGRES_CONN_STR"
	grpcAddr   = ":9099"
	httpAddr   = ":8088"
)

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

	// Open a connection to the database cluster.
	connStr := os.Getenv(connEnvVar)
	postgres, err := sql.Open(driver, connStr)
	if err != nil {
		panic(err)
	}

	defer postgres.Close()
	db := database.New(postgres)

	// Set up the following middlewares on unary/stream requests, (ordering of
	// these matters to some extent):
	// - metrics
	// - auth
	// - rate limiting
	// - logging
	// - req validation
	// - tracing
	// - custom
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			openmetrics.UnaryServerInterceptor(metrics),
			auth.UnaryServerInterceptor(server.Authorize),
			ratelimit.UnaryServerInterceptor(limiter),
			recovery.UnaryServerInterceptor(recoveryOpts...),
			server.CustomUnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			openmetrics.StreamServerInterceptor(metrics),
			auth.StreamServerInterceptor(server.Authorize),
			ratelimit.StreamServerInterceptor(limiter),
			recovery.StreamServerInterceptor(recoveryOpts...),
			server.CustomStreamInterceptor,
		),
	)
	httpServer := &http.Server{Addr: httpAddr}

	metrics.InitializeMetrics(grpcServer)
	ctx := context.Background()

	// Register service RPCs on server
	server.RegisterPostgresUserService(ctx, grpcServer, db)
	server.RegisterPostgresProductService(ctx, grpcServer)

	// Enable grpc reflection for grpcurl
	reflection.Register(grpcServer)

	g := &run.Group{}
	g.Add(server.ServeGrpc(grpcServer, grpcAddr))
	g.Add(server.ServeHttp(httpServer, registry))

	if err = g.Run(); err != nil {
		os.Exit(1)
	}
}
