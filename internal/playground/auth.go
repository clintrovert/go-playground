package playground

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Authorize(ctx context.Context) (context.Context, error) {
	token, err := auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}
	// TODO: This is example only, perform proper Oauth/OIDC verification!
	if token != "test" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token")
	}

	// NOTE: You can also pass the token in the context for further interceptors
	// or gRPC service code.
	return ctx, nil
}
