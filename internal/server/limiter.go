package server

import "context"

type RateLimiter struct{}

func NewRateLimiter() RateLimiter {
	return RateLimiter{}
}

func (rl RateLimiter) Limit(ctx context.Context) error {
	return nil
}
