package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/readactedworks/go-http-server/api/model"
	"github.com/readactedworks/go-http-server/internal/test/mocks"
	"github.com/readactedworks/go-http-server/internal/test/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type testUserService struct {
	service *UserService
	ctx     context.Context
	manager *mocks.MockUserDataManager
}

func newTestUserService(t *testing.T) *testUserService {
	ctrl := gomock.NewController(t)
	manager := mocks.NewMockUserDataManager(ctrl)
	service, _ := NewUserService(manager, logrus.New())

	return &testUserService{
		manager: manager,
		service: service,
		ctx:     context.Background(),
	}
}

func TestGetUser_ValidRequest_ShouldSucceed(t *testing.T) {
	tester := newTestUserService(t)
	expected := utils.GenerateRandomUser()
	request := &model.GetUserRequest{
		UserId: expected.Id,
	}

	tester.manager.EXPECT().
		GetUser(tester.ctx, expected.Id).
		Return(expected, nil).
		Times(1)

	response, err := tester.service.GetUser(tester.ctx, request)
	assert.NoError(t, err)
	assertUserEqual(t, expected, response.User)
}

func TestGetUser_InvalidRequest_ShouldError(t *testing.T) {
	tester := newTestUserService(t)
	request := &model.GetUserRequest{
		UserId: "",
	}

	response, err := tester.service.GetUser(tester.ctx, request)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestGetUser_DbGetUserError_ShouldError(t *testing.T) {
	tester := newTestUserService(t)
	expected := utils.GenerateRandomUser()
	request := &model.GetUserRequest{
		UserId: expected.Id,
	}

	tester.manager.EXPECT().
		GetUser(tester.ctx, expected.Id).
		Return(nil, errors.New("test-error")).
		Times(1)

	response, err := tester.service.GetUser(tester.ctx, request)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestCreateUser_ValidRequest_ShouldSucceed(t *testing.T) {
	tester := newTestUserService(t)
	expected := utils.GenerateRandomUser()
	expected.Id = ""
	request := &model.CreateUserRequest{
		Name:     expected.Name,
		Email:    expected.Email,
		Password: expected.Password,
	}

	tester.manager.EXPECT().
		CreateUser(tester.ctx, expected).
		Return(nil).
		Times(1)

	_, err := tester.service.CreateUser(tester.ctx, request)
	assert.NoError(t, err)
}

func assertUserEqual(t *testing.T, expected *model.User, actual *model.User) {
	assert.Equal(t, expected.Id, actual.Id)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Email, actual.Email)
	assert.Equal(t, expected.Password, actual.Password)
}
