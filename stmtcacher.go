package squirrel

import "database/sql"

type preparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

type stmtCacher struct {
	prep  preparer
	cache map[string]*sql.Stmt
}

func NewStmtCacher(prep preparer) *stmtCacher {
	return &stmtCacher{prep: prep, cache: make(map[string]*sql.Stmt)}
}

func (sc *stmtCacher) Prepare(query string) (*sql.Stmt, error) {
	stmt, ok := sc.cache[query]
	if ok {
		return stmt, nil
	}
	stmt, err := sc.prep.Prepare(query)
	if err == nil {
		sc.cache[query] = stmt
	}
	return stmt, err
}

func (sc *stmtCacher) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	stmt, err := sc.Prepare(query)
	if err != nil {
		return
	}
	return stmt.Exec(args...)
}

func (sc *stmtCacher) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	stmt, err := sc.Prepare(query)
	if err != nil {
		return
	}
	return stmt.Query(args...)
}

func (sc *stmtCacher) QueryRow(query string, args ...interface{}) RowScanner {
	stmt, err := sc.Prepare(query)
	if err != nil {
		return &Row{err: err}
	}
	return stmt.QueryRow(args...)
}
