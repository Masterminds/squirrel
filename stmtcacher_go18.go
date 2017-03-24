// +build go1.8

package squirrel

import (
	"context"
	"database/sql"
)

func (sc *stmtCacher) ExecContext(ctx context.Context, query string, args ...interface{}) (res sql.Result, err error) {
	stmt, err := sc.Prepare(query)
	if err != nil {
		return
	}
	return stmt.ExecContext(ctx, args...)
}

func (sc *stmtCacher) QueryContext(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
	stmt, err := sc.Prepare(query)
	if err != nil {
		return
	}
	return stmt.QueryContext(ctx, args...)
}

func (sc *stmtCacher) QueryRowContext(ctx context.Context, query string, args ...interface{}) RowScanner {
	stmt, err := sc.Prepare(query)
	if err != nil {
		return &Row{err: err}
	}
	return stmt.QueryRowContext(ctx, args...)
}
