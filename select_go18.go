// +build go1.8

package squirrel

import (
	"context"
	"database/sql"

	"github.com/lann/builder"
)

func (d *selectData) ExecContext(ctx context.Context) (sql.Result, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return ExecContextWith(ctx, d.RunWith, d)
}

func (d *selectData) QueryContext(ctx context.Context) (*sql.Rows, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return QueryContextWith(ctx, d.RunWith, d)
}

func (d *selectData) QueryRowContext(ctx context.Context) RowScanner {
	if d.RunWith == nil {
		return &Row{err: RunnerNotSet}
	}
	queryRower, ok := d.RunWith.(QueryRower)
	if !ok {
		return &Row{err: RunnerNotQueryRunner}
	}
	return QueryRowContextWith(ctx, queryRower, d)
}

// ExecContext builds and ExecContexts the query with the Runner set by RunWith.
func (b SelectBuilder) ExecContext(ctx context.Context) (sql.Result, error) {
	data := builder.GetStruct(b).(selectData)
	return data.ExecContext(ctx)
}

// QueryContext builds and QueryContexts the query with the Runner set by RunWith.
func (b SelectBuilder) QueryContext(ctx context.Context) (*sql.Rows, error) {
	data := builder.GetStruct(b).(selectData)
	return data.QueryContext(ctx)
}

// QueryRowContext builds and QueryRowContexts the query with the Runner set by RunWith.
func (b SelectBuilder) QueryRowContext(ctx context.Context) RowScanner {
	data := builder.GetStruct(b).(selectData)
	return data.QueryRowContext(ctx)
}
