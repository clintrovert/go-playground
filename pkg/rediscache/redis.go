package rediscache

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type RedisCache struct {
}

func NewRedisCache() *RedisCache {
	return &RedisCache{}
}

func (r *RedisCache) Get(ctx context.Context, key string) (any, bool) {
	return nil, false
}

func (r *RedisCache) Set(
	ctx context.Context,
	key string,
	val any,
	ttl time.Duration,
) error {
	return nil
}

func GenerateRedisKey(
	ctx context.Context,
	req proto.Message,
	info *grpc.UnaryServerInfo,
) (string, error) {
	return "", nil
}
