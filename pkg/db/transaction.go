package db

import (
	"context"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/jackc/pgx/v5"
)

type Transaction struct {
	timeStamp time.Time
	tx        pgx.Tx
	err       error
}

func (ti *Transaction) Close(ctx context.Context) {
	if ti.tx == nil {
		return
	}

	if ti.err != nil {
		// The transaction interface does not return an error on Close. This means it
		// it meaningless to check this error because it is not possible to propagate
		// it back to the caller. This is an acceptable design choice: the transaction
		// already failed and rolling it back is the best effort strategy. If this
		// fails too, there's not much which can be done.
		// nolint: errcheck
		ti.tx.Rollback(ctx)
	} else {
		// See reasoning above: as nothing is returned from this function it is not
		// worth it to check the error here
		// nolint: errcheck
		ti.tx.Commit(ctx)
	}

	ti.tx = nil
}

func (ti *Transaction) TimeStamp() time.Time {
	return ti.timeStamp
}

func (ti *Transaction) Exec(ctx context.Context, sql string, arguments ...any) (int64, error) {
	if ti.tx == nil {
		return int64(0), ErrAlreadyCommitted
	}

	tag, err := ti.tx.Exec(ctx, sql, arguments...)
	ti.updateErrorStatus(err)

	if err != nil {
		return tag.RowsAffected(), analyzeAndWrapDatabaseError(err)
	}

	return tag.RowsAffected(), err
}

func (ti *Transaction) Rollback() error {
	if ti.tx == nil {
		return ErrAlreadyCommitted
	}

	ti.err = errors.FromCode(errForcedRollback)

	return nil
}

func (ti *Transaction) Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error) {
	if ti.tx == nil {
		return nil, ErrAlreadyCommitted
	}

	rows, err := ti.tx.Query(ctx, sql, arguments...)
	ti.updateErrorStatus(err)

	return rows, err
}

func (t *Transaction) updateErrorStatus(err error) {
	if err != nil {
		t.err = err
	}
}
