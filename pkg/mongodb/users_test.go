package mongodb

//const (
//	hexId = "6475496c5bda79c75aab1666"
//)
//
//type mongoDbTestClient struct {
//	database   *UserDatabase
//	collection *mocks.MockMongoCollector
//}
//
//func setupMongoTestClient(t *testing.T) *mongoDbTestClient {
//	ctrl := gomock.NewController(t)
//	coll := mocks.NewMockMongoCollector(ctrl)
//	return &mongoDbTestClient{
//		database:   &UserDatabase{Database{collection: coll}},
//		collection: coll,
//	}
//}
//
//func TestCreateUser_ValidArgs_ShouldSucceed(t *testing.T) {
//	tester := setupMongoTestClient(t)
//
//	tester.collection.EXPECT().
//		InsertOne(gomock.Any(), gomock.Any()).
//		Return(nil, nil)
//
//	err := tester.database.CreateUser(nil, utils.GenerateRandomUser())
//	assert.NoError(t, err)
//}

// Needs to be wrapped for Decoding
//func TestGetUser_ValidArgs_ShouldSucceed(t *testing.T) {
//	tester := setupMongoTestClient(t)
//	expected := utils.GenerateRandomUser()
//
//	tester.collection.EXPECT().
//		FindOne(gomock.Any(), gomock.Any()).
//		Return()
//
//	actual, err := tester.database.GetUser(context.Background(), hexId)
//	assert.Equal(t, expected.Id, actual.Id)
//	assert.NoError(t, err)
//}
