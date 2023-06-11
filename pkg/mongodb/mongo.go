package mongodb

import "github.com/clintrovert/go-playground/pkg/mongodb/collection"

type Database struct {
	collection collection.MongoCollector
}
