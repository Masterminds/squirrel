// +build go1.8

package squirrel

import (
	"context"
	"database/sql"

	"github.com/lann/builder"
)

func (d *insertData) ExecContext(ctx context.Context) (sql.Result, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	if d.SerializeWith == nil {
		return nil, SerializerNotSet
	}
	ctxRunner, ok := d.RunWith.(ExecerContext)
	if !ok {
		return nil, NoContextSupport
	}
	return ExecContextWith(ctx, ctxRunner, d, d.SerializeWith)
}

func (d *insertData) QueryContext(ctx context.Context) (*sql.Rows, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	if d.SerializeWith == nil {
		return nil, SerializerNotSet
	}
	ctxRunner, ok := d.RunWith.(QueryerContext)
	if !ok {
		return nil, NoContextSupport
	}
	return QueryContextWith(ctx, ctxRunner, d, d.SerializeWith)
}

func (d *insertData) QueryRowContext(ctx context.Context) RowScanner {
	if d.RunWith == nil {
		return &Row{err: RunnerNotSet}
	}
	if d.SerializeWith == nil {
		return &Row{err: SerializerNotSet}
	}
	queryRower, ok := d.RunWith.(QueryRowerContext)
	if !ok {
		if _, ok := d.RunWith.(QueryerContext); !ok {
			return &Row{err: RunnerNotQueryRunner}
		}
		return &Row{err: NoContextSupport}
	}
	return QueryRowContextWith(ctx, queryRower, d, d.SerializeWith)
}

// ExecContext builds and ExecContexts the query with the Runner set by RunWith.
func (b InsertBuilder) ExecContext(ctx context.Context) (sql.Result, error) {
	data := builder.GetStruct(b).(insertData)
	return data.ExecContext(ctx)
}

// QueryContext builds and QueryContexts the query with the Runner set by RunWith.
func (b InsertBuilder) QueryContext(ctx context.Context) (*sql.Rows, error) {
	data := builder.GetStruct(b).(insertData)
	return data.QueryContext(ctx)
}

// QueryRowContext builds and QueryRowContexts the query with the Runner set by RunWith.
func (b InsertBuilder) QueryRowContext(ctx context.Context) RowScanner {
	data := builder.GetStruct(b).(insertData)
	return data.QueryRowContext(ctx)
}

// ScanContext is a shortcut for QueryRowContext().Scan.
func (b InsertBuilder) ScanContext(ctx context.Context, dest ...interface{}) error {
	return b.QueryRowContext(ctx).Scan(dest...)
}
