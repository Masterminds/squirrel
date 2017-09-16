// +build go1.8

package squirrel

import (
	"context"
	"database/sql"
)

// PrepareerContext is the interface that wraps the Prepare and PrepareContext methods.
//
// Prepare executes the given query as implemented by database/sql.Prepare.
// PrepareContext executes the given query as implemented by database/sql.PrepareContext.
type PreparerContext interface {
	Preparer
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// DBProxyContext groups the Execer, Queryer, QueryRower and PreparerContext interfaces.
type DBProxyContext interface {
	Execer
	Queryer
	QueryRower
	PreparerContext
}

// NewStmtCacher returns a DBProxy wrapping prep that caches Prepared Stmts.
//
// Stmts are cached based on the string value of their queries.
func NewStmtCacher(prep PreparerContext) DBProxyContext {
	return &stmtCacher{prep: prep, cache: make(map[string]*sql.Stmt)}
}

func (sc *stmtCacher) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	ctxPrep, ok := sc.prep.(PreparerContext)
	if !ok {
		return nil, NoContextSupport
	}
	sc.mu.RLock()
	stmt, ok := sc.cache[query]
	sc.mu.RUnlock()
	if ok {
		return stmt, nil
	}
	stmt, err := ctxPrep.PrepareContext(ctx, query)
	if err == nil {
		sc.mu.Lock()
		sc.cache[query] = stmt
		sc.mu.Unlock()
	}
	return stmt, err
}

func (sc *stmtCacher) ExecContext(ctx context.Context, query string, args ...interface{}) (res sql.Result, err error) {
	stmt, err := sc.PrepareContext(ctx, query)
	if err != nil {
		return
	}
	return stmt.ExecContext(ctx, args...)
}

func (sc *stmtCacher) QueryContext(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
	stmt, err := sc.PrepareContext(ctx, query)
	if err != nil {
		return
	}
	return stmt.QueryContext(ctx, args...)
}

func (sc *stmtCacher) QueryRowContext(ctx context.Context, query string, args ...interface{}) RowScanner {
	stmt, err := sc.PrepareContext(ctx, query)
	if err != nil {
		return &Row{err: err}
	}
	return stmt.QueryRowContext(ctx, args...)
}
