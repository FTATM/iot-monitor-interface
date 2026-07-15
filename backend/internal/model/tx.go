package model

import "context"

// Transaction gives the Service manual control over the DB transaction
type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Context() context.Context // Returns the context with the hidden tx
}

// TransactionManager is used by the Service to start a transaction
type TransactionManager interface {
	Begin(ctx context.Context) (Transaction, error)
}
