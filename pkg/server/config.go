package server

import (
	"time"

	"github.com/clintrovert/go-playground/pkg/cache"
	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/prometheus/client_golang/prometheus"
)

type timeoutInterceptorConfig struct {
	timeout time.Duration
}

type rateLimitInterceptorConfig struct {
	limiter ratelimit.Limiter
}

type metricsInterceptorConfig struct {
	metrics  *openmetrics.ServerMetrics
	registry *prometheus.Registry
}

type authInterceptorConfig struct {
	authFunc auth.AuthFunc
}

type recoveryInterceptorConfig struct {
	opts []recovery.Option
}

type cacheInterceptorConfig struct {
	kvc    cache.KeyValCache
	keyGen cache.KeyGenerationFunc
	ttl    time.Duration
}
