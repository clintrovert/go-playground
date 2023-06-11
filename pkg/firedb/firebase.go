package firedb

import (
	"github.com/readactedworks/go-http-server/pkg/firedb/references"
)

// Database provides access to specific actions in Firebase real-time
// data store.
type Database struct {
	referenceCreator  references.Creator
	referenceOperator references.OperatorCreator
}
