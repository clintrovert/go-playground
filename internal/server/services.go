package server

import (
	"context"
	"fmt"

	"github.com/clintrovert/go-playground/api/model"
	v1 "github.com/clintrovert/go-playground/api/v1"
	"github.com/clintrovert/go-playground/internal/postgres/database"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	mongoDbUrlEnvVar = "MONGO_DB_URL"
)

func RegisterPostgresUserService(
	ctx context.Context,
	server *grpc.Server,
	queries *database.Queries,
) {
	svc, err := v1.NewUserService(queries, logrus.New())
	if err != nil {
		panic(fmt.Sprintf("user service failed initialization - " + err.Error()))
	}
	model.RegisterUserServiceServer(server, svc)
	logrus.Info("user service registered")
}

func RegisterPostgresProductService(ctx context.Context, server *grpc.Server) {
	logrus.Info("product service registered")
}

//func RegisterMongoUserService(ctx context.Context, server *grpc.Server) {
//	client, err := mongo.Connect(
//		ctx,
//		options.Client().ApplyURI(os.Getenv(mongoDbUrlEnvVar)),
//	)
//	if err != nil {
//		log.Fatalln("error in initializing mongo client: ", err)
//	}
//	collection := client.Database("playground").Collection("users")
//	user := mongodb.NewUserDatabase(collection)
//	svc, err := v1.NewUserService(user, logrus.New())
//	if err != nil {
//		panic(fmt.Sprintf("user service failed initialization - " + err.Error()))
//	}
//	model.RegisterUserServiceServer(server, svc)
//	logrus.Info("user service registered")
//}
//
//func RegisterMongoProductService(ctx context.Context, server *grpc.Server) {
//	logrus.Info("product service registered")
//}
//
//func RegisterFirebaseUserService(
//	ctx context.Context,
//	server *grpc.Server,
//	database *db.Client,
//) {
//	users := firedb.NewUserDatabase(database)
//	svc, err := v1.NewUserService(users, logrus.New())
//	if err != nil {
//		panic(fmt.Sprintf("user service failed initialization - " + err.Error()))
//	}
//
//	model.RegisterUserServiceServer(server, svc)
//	logrus.Info("user service registered")
//}
//
//func RegisterFirebaseProductService(ctx context.Context, server *grpc.Server) {
//	logrus.Info("product service registered")
//}
//
//func RegisterMySqlUserService(ctx context.Context, server *grpc.Server) {
//	logrus.Info("user service registered")
//}
//
//func RegisterMySqlProductService(ctx context.Context, server *grpc.Server) {
//	logrus.Info("product service registered")
//}
