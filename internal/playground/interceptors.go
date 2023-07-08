package playground

import (
	"context"

	"google.golang.org/grpc"
)

func CustomUnaryInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	// Custom logic goes here.
	return handler(ctx, req)
}

func CustomStreamInterceptor(
	srv any,
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {
	// Custom logic goes here.
	return handler(srv, stream)
}
