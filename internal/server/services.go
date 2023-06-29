package server

import (
	"fmt"

	"github.com/clintrovert/go-playground/api/model"
	v1 "github.com/clintrovert/go-playground/api/v1"
	"github.com/clintrovert/go-playground/pkg/postgres/database"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func RegisterUserService(
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

func RegisterProductService(server *grpc.Server, queries *database.Queries) {
	logrus.Info("product service registered")
}
