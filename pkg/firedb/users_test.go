package firedb

import (
	"context"
	"testing"

	"firebase.google.com/go/db"
	"github.com/golang/mock/gomock"
	"github.com/readactedworks/go-http-server/api/model"
	"github.com/readactedworks/go-http-server/internal/test/mocks"
	"github.com/readactedworks/go-http-server/internal/test/utils"
	"github.com/stretchr/testify/assert"
)

type testUserDatabase struct {
	ctrl               *gomock.Controller
	refCreator         *mocks.MockCreator
	refOperatorCreator *mocks.MockOperatorCreator
	refOperator        *mocks.MockOperator

	db  *UserDatabase
	ctx context.Context
}

func setupTestUserDatabase(t *testing.T) testUserDatabase {
	ctrl := gomock.NewController(t)
	creator := mocks.NewMockCreator(ctrl)
	operator := mocks.NewMockOperator(ctrl)
	opCreator := mocks.NewMockOperatorCreator(ctrl)

	return testUserDatabase{
		refCreator: creator,
		db: &UserDatabase{
			Database: Database{
				referenceCreator:  creator,
				referenceOperator: opCreator,
			},
		},
		refOperator:        operator,
		refOperatorCreator: opCreator,
		ctrl:               ctrl,
		ctx:                context.Background(),
	}
}

func TestGetUser_ValidArgs_ShouldSucceed(t *testing.T) {
	tester := setupTestUserDatabase(t)
	expected := utils.GenerateRandomUser()
	ref := &db.Ref{
		Key:  expected.Id,
		Path: userReferenceFmt,
	}

	var actual model.User

	tester.refCreator.EXPECT().
		NewRef("users/" + expected.Id).
		Return(ref).
		Times(1)

	tester.refOperator.EXPECT().
		Get(tester.ctx, &actual).
		DoAndReturn(func(ctx context.Context, value any) error {
			*value.(*model.User) = *expected
			return nil
		}).
		Times(1)

	tester.refOperatorCreator.EXPECT().
		NewOperator(ref).
		Return(tester.refOperator).
		Times(1)

	response, err := tester.db.GetUser(tester.ctx, expected.Id)
	assert.NoError(t, err)
	assert.Equal(t, expected, response)
}
