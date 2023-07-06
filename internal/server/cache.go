package server

import (
	"context"

	"google.golang.org/grpc"
)

type keyConstructFunc func() error

type KeyValCache interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, val any) error
}

type CacheInterceptor struct {
	cache KeyValCache
}

func (c *CacheInterceptor) UnaryServerInterceptor(
	keyFunc keyConstructFunc,
) grpc.UnaryServerInterceptor {
	return nil
}

func (c *CacheInterceptor) StreamServerInterceptor(
	keyFunc keyConstructFunc,
) grpc.StreamServerInterceptor {
	return nil
}
