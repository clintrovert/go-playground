package gateway

import (
	"log"
	"net"
	"net/http"

	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	tcp             = "tcp"
	metricsEndpoint = "/metrics"
)

type Server struct {
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

func NewServer(grpcAddr, httpAddr string) *Server {
	return &Server{
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

func (srv *Server) WithDefaultMetrics() *Server {
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

func (srv *Server) WithCustomMetrics(registerer prometheus.Registerer) *Server {
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

func (srv *Server) WithRateLimiter(limiter ratelimit.Limiter) *Server {
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

func (srv *Server) WithAuth(af auth.AuthFunc) *Server {
	srv.unaryInterceptors = append(
		srv.unaryInterceptors,
		auth.UnaryServerInterceptor(af),
	)
	srv.authEnabled = true
	return srv
}

func (srv *Server) WithGrpcReflection() *Server {
	srv.reflectionEnabled = true
	return srv
}

func (srv *Server) WithRecovery(opts []recovery.Option) *Server {
	srv.unaryInterceptors = append(
		srv.unaryInterceptors,
		recovery.UnaryServerInterceptor(opts...),
	)
	srv.recoveryEnabled = true
	return srv
}

func (srv *Server) GrpcServer() *grpc.Server {
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

func (srv *Server) ServeGrpc() (execute func() error, interrupt func(error)) {
	if srv.grpcServer == nil {
		srv.GrpcServer()
	}

	return func() error {
			l, err := net.Listen(tcp, srv.grpcAddr)
			if err != nil {
				return err
			}
			return srv.grpcServer.Serve(l)
		}, func(err error) {
			srv.grpcServer.GracefulStop()
			srv.grpcServer.Stop()
		}
}

func (srv *Server) ServeHttp() (execute func() error, interrupt func(error)) {
	httpSrv := &http.Server{Addr: srv.httpAddr}
	return func() error {
			m := http.NewServeMux()
			if srv.customMetricsEnabled || srv.defaultMetricsEnabled {
				m.Handle(metricsEndpoint, promhttp.HandlerFor(
					srv.registry,
					promhttp.HandlerOpts{
						EnableOpenMetrics: true,
					},
				))
			}

			httpSrv.Handler = m
			log.Println("starting http server at " + httpSrv.Addr)
			return httpSrv.ListenAndServe()
		}, func(error) {
			if err := httpSrv.Close(); err != nil {
				log.Fatalf("failed to close http server: %v", err)
			}
		}
}
