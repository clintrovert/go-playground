package v1

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/clintrovert/go-playground/api/model"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	notFound     = "not found"
	userLogField = "user_id"
)

var (
	ErrUserIdMissing       = errors.New("user id was not specified")
	ErrUserEmailMissing    = errors.New("user email was not specified")
	ErrUserEmailInvalid    = errors.New("user email was invalid")
	ErrUserNameMissing     = errors.New("user name was not specified")
	ErrUserDeletionFailed  = errors.New("user deletion failed")
	ErrUserPasswordMissing = errors.New("user password was not specified")
)

// UserDataManager provides basic CRUD database operations for Users.
type UserDataManager interface {
	GetUser(ctx context.Context, id string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
}

// UserService provides functionality to manage Users.
type UserService struct {
	db  UserDataManager
	log *logrus.Logger

	model.UnimplementedUserServiceServer
}

// NewUserService creates a new instance of a UserService.
func NewUserService(
	db UserDataManager,
	log *logrus.Logger,
) (*UserService, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	if log == nil {
		return nil, errors.New("log is required")
	}
	return &UserService{
		db:  db,
		log: log,
	}, nil
}

// GetUser retrieves a User by their ID from the database.
func (s *UserService) GetUser(
	ctx context.Context,
	request *model.GetUserRequest,
) (*model.GetUserResponse, error) {
	if err := validateContext(ctx); err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if err := validateGetUserRequest(request); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.db.GetUser(ctx, request.UserId)
	if err != nil {
		s.log.
			WithField(userLogField, request.UserId).
			Error(err)

		if strings.Contains(err.Error(), notFound) {
			return nil, status.Error(codes.NotFound, "")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &model.GetUserResponse{
		User: user,
	}, nil
}

// CreateUser generates a new User in the database.
func (s *UserService) CreateUser(
	ctx context.Context,
	request *model.CreateUserRequest,
) (*model.CreateUserResponse, error) {
	if err := validateContext(ctx); err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if err := validateCreateUserRequest(request); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	encrypted, err := bcrypt.GenerateFromPassword(
		[]byte(strings.TrimSpace(request.Password)),
		bcrypt.DefaultCost,
	)
	if err != nil {
		s.log.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	user := &model.User{
		Name:     strings.TrimSpace(request.Name),
		Email:    strings.TrimSpace(request.Email),
		Password: string(encrypted),
	}

	if err = s.db.CreateUser(ctx, user); err != nil {
		//s.log.
		//	WithField(userLogField, user.Id).
		//	Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &model.CreateUserResponse{Success: true}, nil
}

// UpdateUser modifies an existing User in the database.
func (s *UserService) UpdateUser(
	ctx context.Context,
	request *model.UpdateUserRequest,
) (*model.UpdateUserResponse, error) {
	if err := validateContext(ctx); err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if err := validateUpdateUserRequest(request); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user := &model.User{
		Id:       strings.TrimSpace(request.Id),
		Name:     strings.TrimSpace(request.Name),
		Email:    strings.TrimSpace(request.Email),
		Password: strings.TrimSpace(request.Password),
	}

	if err := s.db.UpdateUser(ctx, user); err != nil {
		s.log.
			WithField(userLogField, user.Id).
			Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &model.UpdateUserResponse{
		Updated: true,
	}, nil
}

// DeleteUser removes an existing User from the database.
func (s *UserService) DeleteUser(
	ctx context.Context,
	request *model.DeleteUserRequest,
) (*model.DeleteUserResponse, error) {
	if err := validateContext(ctx); err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	if err := validateDeleteUserRequest(request); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.db.DeleteUser(ctx, request.UserId); err != nil {
		s.log.
			WithField(userLogField, request.UserId).
			Error(err)
		return nil, status.Error(codes.Internal, ErrUserDeletionFailed.Error())
	}

	return &model.DeleteUserResponse{Deleted: true}, nil
}

func validateGetUserRequest(request *model.GetUserRequest) error {
	if strings.TrimSpace(request.UserId) == "" {
		return ErrUserIdMissing
	}

	return nil
}

func validateCreateUserRequest(request *model.CreateUserRequest) error {
	if strings.TrimSpace(request.Name) == "" {
		return ErrUserNameMissing
	}
	if strings.TrimSpace(request.Email) == "" {
		return ErrUserEmailMissing
	}
	if _, err := mail.ParseAddress(
		strings.TrimSpace(request.Email),
	); err != nil {
		return ErrUserEmailInvalid
	}
	if strings.TrimSpace(request.Password) == "" {
		return ErrUserPasswordMissing
	}

	return nil
}

func validateUpdateUserRequest(request *model.UpdateUserRequest) error {
	if strings.TrimSpace(request.Id) == "" {
		return ErrUserIdMissing
	}
	if strings.TrimSpace(request.Name) == "" {
		return ErrUserNameMissing
	}
	if strings.TrimSpace(request.Email) == "" {
		return ErrUserEmailMissing
	}
	if _, err := mail.ParseAddress(
		strings.TrimSpace(request.Email),
	); err != nil {
		return ErrUserEmailInvalid
	}
	if strings.TrimSpace(request.Password) == "" {
		return ErrUserPasswordMissing
	}
	return nil
}

func validateDeleteUserRequest(request *model.DeleteUserRequest) error {
	if strings.TrimSpace(request.UserId) == "" {
		return ErrUserIdMissing
	}

	return nil
}

func validateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context is required")
	}
	return nil
}
