package mongodb

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/readactedworks/go-http-server/internal/test/mocks"
	"github.com/readactedworks/go-http-server/internal/test/utils"
	"github.com/stretchr/testify/assert"
)

type mongoDbTestClient struct {
	database   *UserDatabase
	collection *mocks.MockMongoCollector
}

func setupMongoTestClient(t *testing.T) *mongoDbTestClient {
	ctrl := gomock.NewController(t)
	coll := mocks.NewMockMongoCollector(ctrl)
	return &mongoDbTestClient{
		database:   &UserDatabase{Database{collection: coll}},
		collection: coll,
	}
}

func TestCreateUser_ValidArgs_ShouldSucceed(t *testing.T) {
	tester := setupMongoTestClient(t)

	tester.collection.EXPECT().
		InsertOne(gomock.Any(), gomock.Any()).
		Return(nil, nil)

	err := tester.database.CreateUser(nil, utils.GenerateRandomUser())
	assert.NoError(t, err)
}
