// +build go1.8

package squirrel

import (
	"context"
	"database/sql"
)

func (s *DBStub) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	s.LastExecSql = query
	s.LastExecArgs = args
	return nil, nil
}

func (s *DBStub) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	s.LastQuerySql = query
	s.LastQueryArgs = args
	return nil, nil
}

func (s *DBStub) QueryRowContext(ctx context.Context, query string, args ...interface{}) RowScanner {
	s.LastQueryRowSql = query
	s.LastQueryRowArgs = args
	return &Row{RowScanner: &RowStub{}}
}
