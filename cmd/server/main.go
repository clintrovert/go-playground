package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/clintrovert/go-playground/api/model"
	"github.com/clintrovert/go-playground/internal/server"
	metrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	excludeAuth := func(ctx context.Context, service string) bool {
		return model.AuthService_ServiceDesc.ServiceName != service
	}

	// Load the certificate and private key files
	certificate, err := tls.LoadX509KeyPair("certs/cert.pem", "certs/key.pem")
	if err != nil {
		log.Fatalf("Failed to load certificate and key: %v", err)
	}

	// Create a certificate pool and add the self-signed certificate
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("certs/cert.pem")
	if err != nil {
		log.Fatalf("Failed to load CA certificate: %v", err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("Failed to append CA certificate")
	}

	// Create the TLS credentials using the certificate and pool
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			metrics.UnaryServerInterceptor(srvMetrics),
			selector.UnaryServerInterceptor(
				auth.UnaryServerInterceptor(server.Authorize),
				selector.MatchFunc(excludeAuth),
			),
			unaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			metrics.StreamServerInterceptor(srvMetrics),
			//auth.StreamServerInterceptor(server.Authorize),
			streamInterceptor,
		),
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	)

	srvMetrics.InitializeMetrics(srv)
	registerUserService(ctx, srv, mongoDb)
	server.RegisterAuthService(ctx, srv)
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
