package squirrel

import (
	"database/sql"
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

type stmtCacher struct {
	prep  Preparer
	cache map[string]*sql.Stmt
	mu    sync.Mutex
}

// NewStmtCacher returns a DBProxy wrapping prep that caches Prepared Stmts.
//
// Stmts are cached based on the string value of their queries.
func NewStmtCacher(prep Preparer) DBProxy {
	return &stmtCacher{prep: prep, cache: make(map[string]*sql.Stmt)}
}

func (sc *stmtCacher) Prepare(query string) (*sql.Stmt, error) {
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

type DBProxyBeginner interface {
	DBProxy
	Begin() (*sql.Tx, error)
}

type stmtCacheProxy struct {
	DBProxy
	db *sql.DB
}

func NewStmtCacheProxy(db *sql.DB) DBProxyBeginner {
	return &stmtCacheProxy{DBProxy: NewStmtCacher(db), db: db}
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
