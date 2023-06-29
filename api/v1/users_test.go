package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/clintrovert/go-playground/api/model"
	"github.com/clintrovert/go-playground/internal/postgres/database"
	"github.com/clintrovert/go-playground/internal/test/mocks"
	"github.com/clintrovert/go-playground/internal/test/utils"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type testUserService struct {
	service  *UserService
	ctx      context.Context
	database *mocks.MockUserDatabase
}

func newTestUserService(t *testing.T) *testUserService {
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockUserDatabase(ctrl)
	service, _ := NewUserService(manager, logrus.New())

	return &testUserService{
		database: manager,
		service:  service,
		ctx:      context.Background(),
	}
}

func TestGetUser_ValidRequest_ShouldSucceed(t *testing.T) {
	tester := newTestUserService(t)
	expected := utils.GenerateRandomUser()
	request := &model.GetUserRequest{
		UserId: expected.UserID,
	}

	tester.database.EXPECT().
		GetUser(tester.ctx, expected.UserID).
		Return(expected, nil).
		Times(1)

	response, err := tester.service.GetUser(tester.ctx, request)
	assert.NoError(t, err)
	assertUserEqual(t, expected, response.User)
}

func TestGetUser_InvalidRequest_ShouldError(t *testing.T) {
	tester := newTestUserService(t)
	request := &model.GetUserRequest{
		UserId: -1,
	}

	response, err := tester.service.GetUser(tester.ctx, request)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGetUser_DbGetUserError_ShouldError(t *testing.T) {
	tester := newTestUserService(t)
	expected := utils.GenerateRandomUser()
	request := &model.GetUserRequest{
		UserId: expected.UserID,
	}

	tester.database.EXPECT().
		GetUser(tester.ctx, expected.UserID).
		Return(database.User{}, errors.New("test-error")).
		Times(1)

	response, err := tester.service.GetUser(tester.ctx, request)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestCreateUser_ValidRequest_ShouldSucceed(t *testing.T) {
	tester := newTestUserService(t)
	expected := utils.GenerateRandomUser()
	expected.UserID = 1

	request := &model.CreateUserRequest{
		Name:     expected.Name.String,
		Email:    expected.Email.String,
		Password: expected.Password.String,
	}

	tester.database.EXPECT().
		CreateUser(tester.ctx, gomock.Any()).
		Return(nil).
		Times(1)

	_, err := tester.service.CreateUser(tester.ctx, request)
	assert.NoError(t, err)
}

func assertUserEqual(t *testing.T, expected database.User, actual *model.User) {
	assert.Equal(t, expected.UserID, actual.Id)
	assert.Equal(t, expected.Name.String, actual.Name)
	assert.Equal(t, expected.Email.String, actual.Email)
}
