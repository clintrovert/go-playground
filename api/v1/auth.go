package v1

import (
	"context"

	"github.com/clintrovert/go-playground/api/model"
)

type AuthService struct {
	model.UnimplementedAuthServiceServer
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) GenerateBearerToken(
	ctx context.Context,
	request *model.GenerateBearerTokenRequest,
) (*model.GenerateBearerTokenResponse, error) {
	return nil, nil
}
