package main

import (
	"context"
	"log"
	"net/http"
	"os"

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

// This is for the purposes of demo. Remove in individual implementation.
type databaseType string

const (
	mysql    databaseType = "mysql"
	mongo    databaseType = "mongodb"
	firebase databaseType = "firebase"
	postgres databaseType = "postgres"

	grpcAddr = ":9099"
	httpAddr = ":8088"
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

	// Set up the following middlewares on unary/stream RPC requests:
	// - metrics
	// - auth
	// - rate limiting
	// - logging
	// - req validation
	// - tracing
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			openmetrics.UnaryServerInterceptor(metrics),
			auth.UnaryServerInterceptor(server.Authorize),
			ratelimit.UnaryServerInterceptor(limiter),
			recovery.UnaryServerInterceptor(recoveryOpts...),
			customUnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			openmetrics.StreamServerInterceptor(metrics),
			auth.StreamServerInterceptor(server.Authorize),
			ratelimit.StreamServerInterceptor(limiter),
			recovery.StreamServerInterceptor(recoveryOpts...),
			customStreamInterceptor,
		),
	)
	httpServer := &http.Server{Addr: httpAddr}

	metrics.InitializeMetrics(grpcServer)
	ctx := context.Background()

	// Register service RPCs on server
	registerUserService(ctx, grpcServer, postgres)
	registerProductService(ctx, grpcServer, postgres)

	// Enable grpc reflection for grpcurl
	reflection.Register(grpcServer)

	g := &run.Group{}
	g.Add(server.ServeGrpc(grpcServer, grpcAddr))
	g.Add(server.ServeHttp(httpServer, registry))

	if err := g.Run(); err != nil {
		os.Exit(1)
	}
}

func customUnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	// Custom logic goes here.
	return handler(ctx, req)
}

func customStreamInterceptor(
	srv any,
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {
	// Custom logic goes here.
	return handler(srv, stream)
}

// This is for the purposes of demo. Remove in individual implementation.
func registerUserService(
	ctx context.Context,
	srv *grpc.Server,
	databaseType databaseType,
) {
	switch databaseType {
	case postgres:
		server.RegisterPostgresUserService(ctx, srv)
	case mysql:
		server.RegisterMySqlUserService(ctx, srv)
	case firebase:
		server.RegisterFirebaseUserService(ctx, srv)
	case mongo:
		server.RegisterMongoUserService(ctx, srv)
	default:
		log.Fatalf("database type %s not supported", databaseType)
	}
}

// This is for the purposes of demo. Remove in individual implementation.
func registerProductService(
	ctx context.Context,
	srv *grpc.Server,
	databaseType databaseType,
) {
	switch databaseType {
	case postgres:
		server.RegisterPostgresProductService(ctx, srv)
	case mysql:
		server.RegisterMySqlProductService(ctx, srv)
	case firebase:
		server.RegisterFirebaseProductService(ctx, srv)
	case mongo:
		server.RegisterMongoProductService(ctx, srv)
	default:
		log.Fatalf("database type %s not supported", databaseType)
	}
}
