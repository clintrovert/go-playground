package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/clintrovert/go-playground/internal/server"
	metrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type databaseType string

const (
	mysql      databaseType = "mysql"
	postgres   databaseType = "postgres"
	mongoDb    databaseType = "mongodb"
	firebaseDb databaseType = "firebase"

	tcp             = "tcp"
	grpcAddr        = ":9099"
	httpAddr        = ":9095"
	metricsEndpoint = "/metrics"
)

func main() {
	srvMetrics := metrics.NewRegisteredServerMetrics(
		prometheus.DefaultRegisterer,
		metrics.WithServerHandlingTimeHistogram(),
	)
	registry := prometheus.NewRegistry()
	registry.MustRegister(srvMetrics)
	ctx := context.Background()

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			metrics.UnaryServerInterceptor(srvMetrics),
			//auth.UnaryServerInterceptor(server.Authorize),
			unaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			metrics.StreamServerInterceptor(srvMetrics),
			//auth.StreamServerInterceptor(server.Authorize),
			streamInterceptor,
		),
	)

	srvMetrics.InitializeMetrics(srv)
	registerUserService(ctx, srv, mongoDb)
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

func unaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	logrus.Info(info.FullMethod + " requested")
	return handler(ctx, req)
}

func streamInterceptor(
	srv any,
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {
	logrus.Info(info.FullMethod + " requested")
	return handler(srv, stream)
}

func registerUserService(
	ctx context.Context,
	srv *grpc.Server,
	databaseType databaseType,
) {
	switch databaseType {
	case firebaseDb:
		server.RegisterFirebaseUserService(ctx, srv)
	case mongoDb:
		server.RegisterMongoUserService(ctx, srv)
	case mysql:
		log.Fatalf("database type %s not supported", databaseType)
	case postgres:
		log.Fatalf("database type %s not supported", databaseType)
	default:
		log.Fatalf("database type %s not supported", databaseType)
	}
}
