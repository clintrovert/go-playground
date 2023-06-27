package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/clintrovert/go-playground/internal/server"
	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
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

	tcp             = "tcp"
	grpcAddr        = ":9099"
	httpAddr        = ":8088"
	metricsEndpoint = "/metrics"
)

func main() {
	// Setup prometheus register/registry.
	metrics := openmetrics.NewRegisteredServerMetrics(
		prometheus.DefaultRegisterer,
		openmetrics.WithServerHandlingTimeHistogram(),
	)
	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics)

	limiter := server.NewRateLimiter()
	// Set up the following middlewares on unary/stream RPC requests:
	// - metrics
	// - auth
	// - rate limiting
	// - logging
	// - tracing
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			openmetrics.UnaryServerInterceptor(metrics),
			auth.UnaryServerInterceptor(server.Authorize),
			ratelimit.UnaryServerInterceptor(limiter),
			customUnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			openmetrics.StreamServerInterceptor(metrics),
			auth.StreamServerInterceptor(server.Authorize),
			ratelimit.StreamServerInterceptor(limiter),
			customStreamInterceptor,
		),
	)
	metrics.InitializeMetrics(srv)

	ctx := context.Background()

	// Register service RPCs on server
	registerUserService(ctx, srv, postgres)
	registerProductService(ctx, srv, postgres)

	// Enable grpc reflection for grpcurl
	reflection.Register(srv)

	g := &run.Group{}
	g.Add(func() error {
		l, err := net.Listen(tcp, grpcAddr)
		if err != nil {
			return err
		}
		return srv.Serve(l)
	}, func(err error) {
		srv.GracefulStop()
		srv.Stop()
	})

	// Setup metrics endpoint served over http
	httpSrv := &http.Server{Addr: httpAddr}
	g.Add(
		func() error {
			m := http.NewServeMux()
			m.Handle(metricsEndpoint, promhttp.HandlerFor(
				registry,
				promhttp.HandlerOpts{
					EnableOpenMetrics: true,
				},
			))
			httpSrv.Handler = m
			log.Println("starting http server at " + httpAddr)
			return httpSrv.ListenAndServe()
		}, func(error) {
			if err := httpSrv.Close(); err != nil {
				log.Fatalf("failed to close http server: %v", err)
			}
		})

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
	logrus.Info(info.FullMethod + " requested.")
	return handler(ctx, req)
}

func customStreamInterceptor(
	srv any,
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {
	logrus.Info(info.FullMethod + " requested.")
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
