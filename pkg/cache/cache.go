package cache

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

var cacheHit = metadata.Pairs("x-cache", "hit")

type KeyGenerationFunc func(
	ctx context.Context,
	req proto.Message,
	info *grpc.UnaryServerInfo,
) (string, error)

type KeyValCache interface {
	Get(ctx context.Context, key string) (any, bool)
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
}

type CacheInterceptor struct {
	cache KeyValCache
	log   *logrus.Entry
}

func NewKeyValCacheInterceptor(
	cache KeyValCache,
	log *logrus.Entry,
) *CacheInterceptor {
	return &CacheInterceptor{
		cache: cache,
		log:   log,
	}
}

func (c *CacheInterceptor) UnaryServerInterceptor(
	keyFunc KeyGenerationFunc, ttl time.Duration,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		request any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		key, err := keyFunc(ctx, request.(proto.Message), info)
		if err != nil {
			return nil, err
		}

		if val, found := c.cache.Get(ctx, key); found {
			if err = grpc.SendHeader(ctx, cacheHit); err != nil {
				return val, nil
			}
		}

		resp, err := handler(ctx, request)
		if err != nil {
			return nil, err
		}
		if err = c.cache.Set(ctx, key, resp, ttl); err != nil {

		}
		return resp, nil
	}
}

func (c *CacheInterceptor) StreamServerInterceptor(
	keyFunc KeyGenerationFunc, ttl time.Duration,
) grpc.StreamServerInterceptor {
	return nil
}
