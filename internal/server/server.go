package server

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"github.com/readactedworks/go-http-server/api/model"
	v1 "github.com/readactedworks/go-http-server/api/v1"
	"github.com/readactedworks/go-http-server/pkg/firedb"
	"github.com/readactedworks/go-http-server/pkg/mongodb"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

const (
	mongoDbUrlEnvVar   = "MONGO_DB_URL"
	firebaseUrlEnvVar  = "FIREBASE_DB_URL"
	firebaseCredEnvVar = "FIREBASE_CRED_FILEPATH"
)

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
