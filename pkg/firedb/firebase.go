package firedb

import (
	"github.com/clintrovert/go-playground/pkg/firedb/references"
)

// Database provides access to specific actions in Firebase real-time
// data store.
type Database struct {
	referenceCreator  references.Creator
	referenceOperator references.OperatorCreator
}
