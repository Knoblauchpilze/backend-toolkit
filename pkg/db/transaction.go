package db

import (
	"context"
	"time"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/db/pgx"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	jpgx "github.com/jackc/pgx/v5"
)

type Transaction interface {
	Close(ctx context.Context)
	TimeStamp() time.Time

	Exec(ctx context.Context, sql string, arguments ...any) (int64, error)
}

type transactionImpl struct {
	timeStamp time.Time
	tx        jpgx.Tx
	err       error
}

func (ti *transactionImpl) Close(ctx context.Context) {
	if ti.err != nil {
		ti.tx.Rollback(ctx)
	} else {
		ti.tx.Commit(ctx)
	}

	ti.tx = nil
}

func (ti *transactionImpl) TimeStamp() time.Time {
	return ti.timeStamp
}

func (ti *transactionImpl) Exec(ctx context.Context, sql string, arguments ...any) (int64, error) {
	if ti.tx == nil {
		return int64(0), errors.NewCode(AlreadyCommitted)
	}

	tag, err := ti.tx.Exec(ctx, sql, arguments...)
	ti.updateErrorStatus(err)

	if err != nil {
		return tag.RowsAffected(), pgx.AnalyzeAndWrapPgError(err)
	}

	return tag.RowsAffected(), err
}

func (ti *transactionImpl) query(ctx context.Context, sql string, arguments ...any) (jpgx.Rows, error) {
	if ti.tx == nil {
		return nil, errors.NewCode(AlreadyCommitted)
	}

	rows, err := ti.tx.Query(ctx, sql, arguments...)
	ti.updateErrorStatus(err)

	return rows, err
}

func (t *transactionImpl) updateErrorStatus(err error) {
	if err != nil {
		t.err = err
	}
}
