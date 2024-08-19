package indexstorage

import (
	"context"
)

type Committable interface {
	// Rollback attempts to remove all local changes and revert to the state where the backing storage is the
	// authoritative source of information.
	Rollback(ctx context.Context) error
	// Commit attempts to write all changes to the backing storage. If it fails, it attempts to make the remaining
	// state on the backing storage consistent, but it may not be able to guarantee it. It will leave the local
	// directory in such a state that it can continue the upload by calling Commit again.
	Commit(ctx context.Context) error

	// Recover attempts to recover a previously-aborted commit if any. If no commit was started, it rolls
	// back any changes.
	Recover(ctx context.Context) error
}

type TransactionalAPI interface {
	API
	Committable
}
