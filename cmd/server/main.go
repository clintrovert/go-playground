package main

import (
	"context"
	"fmt"
	"log"
	"net"

	metrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/readactedworks/go-http-server/internal/server"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	tcp        = "tcp"
	port       = 9099
	firebaseDb = "firebase"
	mysql      = "mysql"
	postgres   = "postgres"
	mongoDb    = "mongodb"
)

func main() {
	var err error
	ctx := context.Background()

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.StreamInterceptor(streamInterceptor),
	)

	registerUserService(ctx, srv, mongoDb)

	//registerServiceMetrics(server, prometheus.DefaultRegisterer)
	reflection.Register(srv)
	lis, err := net.Listen(tcp, fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err = srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func unaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	logrus.Info("unary interceptor")
	return handler(ctx, req)
}

func streamInterceptor(
	srv any,
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {
	return handler(srv, stream)
}

func registerUserService(
	ctx context.Context,
	srv *grpc.Server,
	databaseType string,
) {
	switch databaseType {
	case firebaseDb:
		server.RegisterFirebaseUserService(ctx, srv)
	case mongoDb:
		server.RegisterMongoUserService(ctx, srv)
	case mysql:
		log.Fatalf("database type %s not supported", databaseType)
	default:
		log.Fatalf("database type %s not supported", databaseType)
	}
}

func registerServiceMetrics(server *grpc.Server, reg prometheus.Registerer) {
	srvMetrics := metrics.NewRegisteredServerMetrics(
		reg,
		metrics.WithServerHandlingTimeHistogram(),
	)
	registry := prometheus.NewRegistry()
	registry.MustRegister(srvMetrics)
	srvMetrics.InitializeMetrics(server)
	logrus.Info("service metrics registered")
}
