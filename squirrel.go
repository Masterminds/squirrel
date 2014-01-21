package squirrel

import "database/sql"

type Sqlizer interface {
	ToSql() (string, []interface{}, error)
}

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func ExecWith(db Execer, s Sqlizer) (res sql.Result, err error) {
	query, args, err := s.ToSql()
	if err != nil {
		return
	}
	return db.Exec(query, args...)
}

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func QueryWith(db Queryer, s Sqlizer) (rows *sql.Rows, err error) {
	query, args, err := s.ToSql()
	if err != nil {
		return
	}
	return db.Query(query, args...)
}

type QueryRower interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

func QueryRowWith(db QueryRower, s Sqlizer) *Row {
	query, args, err := s.ToSql()
	return &Row{Row: db.QueryRow(query, args...), sqlErr: err}
}

type Row struct {
	*sql.Row
	sqlErr error
}

func (r *Row) Scan(dest ...interface{}) error {
	if r.sqlErr != nil {
		return r.sqlErr
	}
	return r.Row.Scan(dest...)
}
