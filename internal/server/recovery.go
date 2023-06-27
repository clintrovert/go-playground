package server

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Recover(p any) (err error) {
	return status.Errorf(codes.Unknown, "panic triggered: %v", p)
}
