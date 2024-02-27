package squirrel

import (
	"database/sql"
	"fmt"
	"sync"
)

// Prepareer is the interface that wraps the Prepare method.
//
// Prepare executes the given query as implemented by database/sql.Prepare.
type Preparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

// DBProxy groups the Execer, Queryer, QueryRower, and Preparer interfaces.
type DBProxy interface {
	Execer
	Queryer
	QueryRower
	Preparer
}

// NOTE: NewStmtCache is defined in stmtcacher_ctx.go (Go >= 1.8) or stmtcacher_noctx.go (Go < 1.8).

// StmtCache wraps and delegates down to a Preparer type
//
// It also automatically prepares all statements sent to the underlying Preparer calls
// for Exec, Query and QueryRow and caches the returns *sql.Stmt using the provided
// query as the key. So that it can be automatically re-used.
type StmtCache struct {
	prep  Preparer
	cache map[string]*sql.Stmt
	mu    sync.Mutex
}

// Prepare delegates down to the underlying Preparer and caches the result
// using the provided query as a key
func (sc *StmtCache) Prepare(query string) (*sql.Stmt, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

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

// Exec delegates down to the underlying Preparer using a prepared statement
func (sc *StmtCache) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	stmt, err := sc.Prepare(query)
	if err != nil {
		return
	}
	return stmt.Exec(args...)
}

// Query delegates down to the underlying Preparer using a prepared statement
func (sc *StmtCache) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	stmt, err := sc.Prepare(query)
	if err != nil {
		return
	}
	return stmt.Query(args...)
}

// QueryRow delegates down to the underlying Preparer using a prepared statement
func (sc *StmtCache) QueryRow(query string, args ...interface{}) RowScanner {
	stmt, err := sc.Prepare(query)
	if err != nil {
		return &Row{err: err}
	}
	return stmt.QueryRow(args...)
}

// Clear removes and closes all the currently cached prepared statements
func (sc *StmtCache) Clear() (err error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for key, stmt := range sc.cache {
		delete(sc.cache, key)

		if stmt == nil {
			continue
		}

		if cerr := stmt.Close(); cerr != nil {
			err = cerr
		}
	}

	if err != nil {
		return fmt.Errorf("one or more Stmt.Close failed; last error: %v", err)
	}

	return
}

type DBProxyBeginner interface {
	DBProxy
	Begin() (*sql.Tx, error)
}

type stmtCacheProxy struct {
	DBProxy
	db *sql.DB
}

func NewStmtCacheProxy(db *sql.DB) DBProxyBeginner {
	return &stmtCacheProxy{DBProxy: NewStmtCache(db), db: db}
}

func (sp *stmtCacheProxy) Begin() (*sql.Tx, error) {
	return sp.db.Begin()
}

// DBTransactionProxy wraps transaction and includes DBProxy interface
type DBTransactionProxy interface {
	DBProxy
	Begin() error
	Commit() error
	Rollback() error
}

type stmtCacheTransactionProxy struct {
	DBProxy
	db          *sql.DB
	transaction *sql.Tx
}

// NewStmtCacheTransactionProxy returns a DBTransactionProxy
// wrapping an open transaction in stmtCacher.
// You should use Begin() each time you want a new transaction and
// cache will be valid only for that transaction.
// By default without calling Begin proxy will use simple stmtCacher
//
// Usage example:
//	proxy := sq.NewStmtCacheTransactionProxy(db)
//	mydb := sq.StatementBuilder.RunWith(proxy)
//	insertUsers := mydb.Insert("users").Columns("name")
//	insertUsers.Values("username1").Exec()
//	insertUsers.Values("username2").Exec()
//	proxy.Commit()
//
//	proxy.Begin()
//	insertPets := mydb.Insert("pets").Columns("name", "username")
//	insertPets.Values("petname1", "username1").Exec()
//	insertPets.Values("petname2", "username1").Exec()
//	proxy.Commit()
func NewStmtCacheTransactionProxy(db *sql.DB) (proxy DBTransactionProxy) {
	return &stmtCacheTransactionProxy{DBProxy: NewStmtCacher(db), db: db}
}

func (s *stmtCacheTransactionProxy) Begin() (err error) {
	tr, err := s.db.Begin()

	if err != nil {
		return
	}

	s.DBProxy = NewStmtCacher(tr)
	s.transaction = tr

	return
}

func (s *stmtCacheTransactionProxy) Commit() error {
	defer s.resetProxy()
	return s.transaction.Commit()
}

func (s *stmtCacheTransactionProxy) Rollback() error {
	defer s.resetProxy()
	return s.transaction.Rollback()
}

func (s *stmtCacheTransactionProxy) resetProxy() {
	s.DBProxy = NewStmtCacher(s.db)
}
