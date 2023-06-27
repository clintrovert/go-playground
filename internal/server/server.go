package server

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"github.com/clintrovert/go-playground/api/model"
	v1 "github.com/clintrovert/go-playground/api/v1"
	"github.com/clintrovert/go-playground/pkg/firedb"
	"github.com/clintrovert/go-playground/pkg/mongodb"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	mongoDbUrlEnvVar   = "MONGO_DB_URL"
	firebaseUrlEnvVar  = "FIREBASE_DB_URL"
	firebaseCredEnvVar = "FIREBASE_CRED_FILEPATH"
)

func Authorize(ctx context.Context) (context.Context, error) {
	token, err := auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}
	// TODO: This is example only, perform proper Oauth/OIDC verification!
	if token != "test" {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token")
	}

	// NOTE: You can also pass the token in the context for further interceptors
	// or gRPC service code.
	return ctx, nil
}

func RegisterPostgresUserService(ctx context.Context, server *grpc.Server) {
	logrus.Info("user service registered")
}

func RegisterPostgresProductService(ctx context.Context, server *grpc.Server) {
	logrus.Info("product service registered")
}

func RegisterMongoUserService(ctx context.Context, server *grpc.Server) {
	client, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(os.Getenv(mongoDbUrlEnvVar)),
	)
	if err != nil {
		log.Fatalln("error in initializing mongo client: ", err)
	}
	collection := client.Database("playground").Collection("users")
	user := mongodb.NewUserDatabase(collection)
	svc, err := v1.NewUserService(user, logrus.New())
	if err != nil {
		panic(fmt.Sprintf("user service failed initialization - " + err.Error()))
	}
	model.RegisterUserServiceServer(server, svc)
	logrus.Info("user service registered")
}

func RegisterMongoProductService(ctx context.Context, server *grpc.Server) {
	logrus.Info("product service registered")
}

func RegisterFirebaseUserService(ctx context.Context, server *grpc.Server) {
	conf := &firebase.Config{DatabaseURL: os.Getenv(firebaseUrlEnvVar)}
	opt := option.WithCredentialsFile(os.Getenv(firebaseCredEnvVar))
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		log.Fatalln("error in initializing firebase app: ", err)
	}

	database, err := app.Database(ctx)
	if err != nil {
		log.Fatalln("error in creating firebase DB client: ", err)
	}
	users := firedb.NewUserDatabase(database)
	svc, err := v1.NewUserService(users, logrus.New())
	if err != nil {
		panic(fmt.Sprintf("user service failed initialization - " + err.Error()))
	}

	model.RegisterUserServiceServer(server, svc)
	logrus.Info("user service registered")
}

func RegisterFirebaseProductService(ctx context.Context, server *grpc.Server) {
	logrus.Info("product service registered")
}

func RegisterMySqlUserService(ctx context.Context, server *grpc.Server) {
	logrus.Info("user service registered")
}

func RegisterMySqlProductService(ctx context.Context, server *grpc.Server) {
	logrus.Info("product service registered")
}
