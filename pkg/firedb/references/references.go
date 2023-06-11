package references

import (
	"context"

	"firebase.google.com/go/db"
)

// Contains all the wrapping logic to mock the firedb implementation
// because the library makes it difficult.

// Creator generates a new db.Ref from a path, permitting operations
// on the reference.
type Creator interface {
	NewRef(path string) *db.Ref
}

// OperatorCreator creates a new Operator from a db.Ref, permitting operations
type OperatorCreator interface {
	NewOperator(ref Operator) Operator
}

// Factory is a wrapper around creating db.Ref wrappers.
type Factory struct{}

// NewOperator creates a wrapper for db.Ref so that it can be mocked.
func (f *Factory) NewOperator(ref Operator) Operator {
	return &OperationClient{
		ref: ref,
	}
}

// Operator specifies the contract for db.Ref so that it can be mocked.
type Operator interface {
	Get(ctx context.Context, v any) error
	Set(ctx context.Context, v any) error
	Delete(ctx context.Context) error
}

// OperationClient wraps the db.Ref so that it can be mocked.
type OperationClient struct {
	ref Operator
}

// Get is a wrapper around db.Ref.Get
func (o *OperationClient) Get(ctx context.Context, v any) error {
	return o.ref.Get(ctx, v)
}

// Set is a wrapper around db.Ref.Set
func (o *OperationClient) Set(ctx context.Context, v any) error {
	return o.ref.Set(ctx, v)
}

// Delete is a wrapper around db.Ref.Delete
func (o *OperationClient) Delete(ctx context.Context) error {
	return o.ref.Delete(ctx)
}
