package firedb

import (
	"context"
	"log"
	"os"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"google.golang.org/api/option"

	"github.com/clintrovert/go-playground/pkg/firedb/references"
)

const (
	firebaseUrlEnvVar  = "FIREBASE_DB_URL"
	firebaseCredEnvVar = "FIREBASE_CRED_FILEPATH"
)

// Database provides access to specific actions in Firebase real-time
// data store.
type Database struct {
	referenceCreator  references.Creator
	referenceOperator references.OperatorCreator
}

// NewDatabase creates a new instance of Database.
func NewDatabase(ctx context.Context) (*db.Client, error) {
	url := os.Getenv(firebaseUrlEnvVar)
	if strings.TrimSpace(url) == "" {
		return nil, nil
	}

	credFile := os.Getenv(firebaseCredEnvVar)
	if strings.TrimSpace(credFile) == "" {
		return nil, nil
	}

	conf := &firebase.Config{DatabaseURL: url}
	opt := option.WithCredentialsFile(credFile)
	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		log.Fatalln("error in initializing firebase app: ", err)
	}

	database, err := app.Database(ctx)
	if err != nil {
		log.Fatalln("error in creating firebase DB client: ", err)
	}
	return database, nil
}
