package server

import (
	"net/http"

	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	tcp             = "tcp"
	metricsEndpoint = "/metrics"
)

type Builder struct {
	grpcAddr   string
	httpAddr   string
	grpcServer *grpc.Server
	httpServer *http.Server

	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	metrics            *openmetrics.ServerMetrics
	registry           *prometheus.Registry

	defaultMetricsEnabled bool
	rateLimitEnabled      bool
	authEnabled           bool
	recoveryEnabled       bool
	reflectionEnabled     bool
	customMetricsEnabled  bool
}

func NewBuilder(grpcAddr, httpAddr string) *Builder {
	return &Builder{
		grpcAddr:              grpcAddr,
		httpAddr:              httpAddr,
		defaultMetricsEnabled: false,
		customMetricsEnabled:  false,
		rateLimitEnabled:      false,
		authEnabled:           false,
		recoveryEnabled:       false,
		reflectionEnabled:     false,
	}
}

func (srv *Builder) WithDefaultMetrics() *Builder {
	if srv.customMetricsEnabled {
		panic("cannot only use default metrics or custom metrics")
	}

	metrics := openmetrics.NewRegisteredServerMetrics(
		prometheus.DefaultRegisterer,
		openmetrics.WithServerHandlingTimeHistogram(),
	)
	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics)
	srv.registry = registry

	srv.unaryInterceptors = append(
		srv.unaryInterceptors,
		openmetrics.UnaryServerInterceptor(metrics),
	)
	srv.streamInterceptors = append(
		srv.streamInterceptors,
		openmetrics.StreamServerInterceptor(metrics),
	)
	srv.defaultMetricsEnabled = true

	return srv
}

func (srv *Builder) WithCustomMetrics(registerer prometheus.Registerer) *Builder {
	if srv.defaultMetricsEnabled {
		panic("cannot only use default metrics or custom metrics")
	}

	metrics := openmetrics.NewRegisteredServerMetrics(
		registerer,
		openmetrics.WithServerHandlingTimeHistogram(),
	)
	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics)
	srv.registry = registry

	srv.unaryInterceptors = append(
		srv.unaryInterceptors,
		openmetrics.UnaryServerInterceptor(metrics),
	)
	srv.streamInterceptors = append(
		srv.streamInterceptors,
		openmetrics.StreamServerInterceptor(metrics),
	)
	srv.customMetricsEnabled = true

	return srv
}

func (srv *Builder) WithRateLimiter(limiter ratelimit.Limiter) *Builder {
	srv.unaryInterceptors = append(
		srv.unaryInterceptors,
		ratelimit.UnaryServerInterceptor(limiter),
	)
	srv.streamInterceptors = append(
		srv.streamInterceptors,
		ratelimit.StreamServerInterceptor(limiter),
	)
	srv.rateLimitEnabled = true
	return srv
}

func (srv *Builder) WithAuth(af auth.AuthFunc) *Builder {
	srv.unaryInterceptors = append(
		srv.unaryInterceptors,
		auth.UnaryServerInterceptor(af),
	)
	srv.authEnabled = true
	return srv
}

func (srv *Builder) WithGrpcReflection() *Builder {
	srv.reflectionEnabled = true
	return srv
}

func (srv *Builder) WithRecovery(opts []recovery.Option) *Builder {
	srv.unaryInterceptors = append(
		srv.unaryInterceptors,
		recovery.UnaryServerInterceptor(opts...),
	)
	srv.recoveryEnabled = true
	return srv
}

func (srv *Builder) Build() *Server {
	// TODO: Implement and validate
	return nil
}

func (srv *Builder) generateGrpcServer() *grpc.Server {
	if srv.grpcServer != nil {
		return srv.grpcServer
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(srv.unaryInterceptors...),
		grpc.ChainStreamInterceptor(srv.streamInterceptors...),
	)

	if srv.defaultMetricsEnabled {
		if srv.metrics == nil {
			panic("metrics not defined")
		}
		srv.metrics.InitializeMetrics(grpcServer)
	}

	if srv.reflectionEnabled {
		reflection.Register(grpcServer)
	}
	srv.grpcServer = grpcServer
	return grpcServer
}
