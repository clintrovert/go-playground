package mongodb

import "github.com/readactedworks/go-http-server/pkg/mongodb/collection"

type Database struct {
	collection collection.MongoCollector
}
