package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/clintrovert/go-playground/pkg/cache"
	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/validator"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	tcp = "tcp"
)

// Builder is a construct
type Builder struct {
	grpcAddr, httpAddr string
	grpcServer         *grpc.Server
	httpServer         *http.Server
	metrics            *metricsInterceptorConfig
	rateLimit          *rateLimitInterceptorConfig
	auth               *authInterceptorConfig
	recovery           *recoveryInterceptorConfig
	cache              *cacheInterceptorConfig
	reflectionEnabled  bool
	validationEnabled  bool
}

// NewBuilder generates a new instance of a server.Builder.
func NewBuilder(grpcAddr, httpAddr string) *Builder {
	return &Builder{
		grpcAddr:          grpcAddr,
		httpAddr:          httpAddr,
		reflectionEnabled: false,
	}
}

// WithMetrics adds metrics interceptors for the supplied registerer.
func (b *Builder) WithMetrics(registerer prometheus.Registerer) *Builder {
	metrics := openmetrics.NewRegisteredServerMetrics(
		registerer,
		openmetrics.WithServerHandlingTimeHistogram(),
	)
	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics)

	b.metrics = &metricsInterceptorConfig{
		metrics:  metrics,
		registry: registry,
	}

	return b
}

func (b *Builder) WithCache(
	kvc cache.KeyValCache,
	keyGenFunc cache.KeyGenerationFunc,
	ttl time.Duration,
) *Builder {
	b.cache = &cacheInterceptorConfig{
		kvc:    kvc,
		keyGen: keyGenFunc,
		ttl:    ttl,
	}

	return b
}

func (b *Builder) WithRateLimiter(limiter ratelimit.Limiter) *Builder {
	b.rateLimit = &rateLimitInterceptorConfig{
		limiter: limiter,
	}

	return b
}

func (b *Builder) WithAuth(af auth.AuthFunc) *Builder {
	b.auth = &authInterceptorConfig{
		authFunc: af,
	}
	return b
}

func (b *Builder) WithGrpcReflection() *Builder {
	b.reflectionEnabled = true
	return b
}

func (b *Builder) WithGrpcValidation() *Builder {
	b.validationEnabled = true
	return b
}

func (b *Builder) WithRecovery(opts []recovery.Option) *Builder {
	b.recovery = &recoveryInterceptorConfig{
		opts: opts,
	}

	return b
}

func (b *Builder) Build() (*Server, error) {
	grpcServer, err := b.generateGrpcServer()
	if err != nil {
		return nil, err
	}

	httpServer, err := b.generateHttpServer()
	if err != nil {
		return nil, err
	}

	return &Server{
		GrpcServer: grpcServer,
		HttpServer: httpServer,
		grpcPort:   b.grpcAddr,
		httpPort:   b.httpAddr,
	}, nil
}

func (b *Builder) generateHttpServer() (*http.Server, error) {
	if err := b.validateHttpConfig(); err != nil {
		return nil, err
	}

	return &http.Server{
		Addr: b.httpAddr,
	}, nil
}

func (b *Builder) generateGrpcServer() (*grpc.Server, error) {
	if err := b.validateGrpcConfig(); err != nil {
		return nil, err
	}

	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	if b.metrics != nil && b.metrics.registry != nil {
		unaryInterceptors = append(
			unaryInterceptors,
			openmetrics.UnaryServerInterceptor(b.metrics.metrics),
		)

		streamInterceptors = append(
			streamInterceptors,
			openmetrics.StreamServerInterceptor(b.metrics.metrics),
		)
	}

	if b.auth != nil && b.auth.authFunc != nil {
		unaryInterceptors = append(
			unaryInterceptors,
			auth.UnaryServerInterceptor(b.auth.authFunc),
		)

		streamInterceptors = append(
			streamInterceptors,
			auth.StreamServerInterceptor(b.auth.authFunc),
		)
	}

	if b.rateLimit != nil && b.rateLimit.limiter != nil {
		unaryInterceptors = append(
			unaryInterceptors,
			ratelimit.UnaryServerInterceptor(b.rateLimit.limiter),
		)

		streamInterceptors = append(
			streamInterceptors,
			ratelimit.StreamServerInterceptor(b.rateLimit.limiter),
		)
	}

	if b.validationEnabled {
		unaryInterceptors = append(
			unaryInterceptors,
			validator.UnaryServerInterceptor(true),
		)

		streamInterceptors = append(
			streamInterceptors,
			validator.StreamServerInterceptor(true),
		)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	if b.reflectionEnabled {
		reflection.Register(grpcServer)
	}

	b.grpcServer = grpcServer
	return grpcServer, nil
}

func (b *Builder) validateGrpcConfig() error {
	if b.metrics != nil && b.metrics == nil {
		return errors.New("metrics registry was not defined")
	}
	return nil
}

func (b *Builder) validateHttpConfig() error {
	return nil
}
