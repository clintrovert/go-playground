package v1

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/clintrovert/go-playground/api/model"
	"github.com/clintrovert/go-playground/pkg/cache"
	database2 "github.com/clintrovert/go-playground/pkg/postgres/database"
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
	ErrUserAuthFailed      = errors.New("user authentication failed")
	ErrUserIdInvalid       = errors.New("user id was not specified")
	ErrUserEmailMissing    = errors.New("user email was not specified")
	ErrUserEmailInvalid    = errors.New("user email was invalid")
	ErrUserNameMissing     = errors.New("user name was not specified")
	ErrUserCreateFailed    = errors.New("user creation failed")
	ErrUserUpdateFailed    = errors.New("user update failed")
	ErrUserDeletionFailed  = errors.New("user deletion failed")
	ErrUserPasswordMissing = errors.New("user password was not specified")
)

// UserDatabase provides database operations for Users.
type UserDatabase interface {
	// GetUser retrieves a User by their ID from the database.
	GetUser(ctx context.Context, id int32) (database2.User, error)
	// CreateUser creates a new User in the database.
	CreateUser(ctx context.Context, params database2.CreateUserParams) error
	// UpdateUser updates an existing User in the database.
	UpdateUser(ctx context.Context, params database2.UpdateUserParams) error
	// DeleteUser deletes a User from the database.
	DeleteUser(ctx context.Context, id int32) error
}

// UserService provides functionality to manage Users.
type UserService struct {
	model.UnimplementedUserServiceServer
	db  UserDatabase
	log *logrus.Logger
	kvc cache.KeyValCache
}

// NewUserService creates a new instance of a UserService.
func NewUserService(db UserDatabase, log *logrus.Logger) (*UserService, error) {
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
		return nil, status.Error(codes.Unauthenticated, ErrUserAuthFailed.Error())
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
		User: &model.User{
			Id:      user.UserID,
			Name:    user.Name.String,
			Email:   user.Email.String,
			IsAdmin: user.IsAdmin.Bool,
		},
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

	user := database2.CreateUserParams{
		Name: sql.NullString{
			String: strings.TrimSpace(request.Name),
			Valid:  true,
		},
		Email: sql.NullString{
			String: strings.TrimSpace(request.Email),
			Valid:  true,
		},
		Password: sql.NullString{
			String: string(encrypted),
			Valid:  true,
		},
		IsAdmin: sql.NullBool{
			Bool:  request.IsAdmin,
			Valid: true,
		},
	}

	if err = s.db.CreateUser(ctx, user); err != nil {
		//s.log.
		//	WithField(userLogField, user.Id).
		//	Error(err)
		return nil, status.Error(codes.Internal, ErrUserCreateFailed.Error())
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

	encrypted, err := bcrypt.GenerateFromPassword(
		[]byte(strings.TrimSpace(request.Password)),
		bcrypt.DefaultCost,
	)
	if err != nil {
		s.log.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	user := database2.UpdateUserParams{
		UserID: request.Id,
		Name: sql.NullString{
			String: strings.TrimSpace(request.Name),
			Valid:  true,
		},
		Email: sql.NullString{
			String: strings.TrimSpace(request.Email),
			Valid:  true,
		},
		Password: sql.NullString{
			String: string(encrypted),
			Valid:  true,
		},
		IsAdmin: sql.NullBool{
			Bool:  request.IsAdmin,
			Valid: true,
		},
	}

	if err = s.db.UpdateUser(ctx, user); err != nil {
		s.log.
			WithField(userLogField, fmt.Sprintf("%v", user.UserID)).
			Error(err)
		return nil, status.Error(codes.Internal, ErrUserUpdateFailed.Error())
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
	if request.UserId < 1 {
		return ErrUserIdInvalid
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
	if request.UserId > 0 {
		return ErrUserIdInvalid
	}

	return nil
}

func validateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context is required")
	}
	return nil
}
