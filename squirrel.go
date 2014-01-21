package squirrel

import "database/sql"

type Sqlizer interface {
	ToSql() (string, []interface{}, error)
}

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type QueryRower interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

type Runner interface {
	Execer
	Queryer
	QueryRower
}

func ExecWith(db Execer, s Sqlizer) (res sql.Result, err error) {
	query, args, err := s.ToSql()
	if err != nil {
		return
	}
	return db.Exec(query, args...)
}

func QueryWith(db Queryer, s Sqlizer) (rows *sql.Rows, err error) {
	query, args, err := s.ToSql()
	if err != nil {
		return
	}
	return db.Query(query, args...)
}

func QueryRowWith(db QueryRower, s Sqlizer) *Row {
	query, args, err := s.ToSql()
	return &Row{Row: db.QueryRow(query, args...), err: err}
}

type Row struct {
	*sql.Row
	err error
}

func (r *Row) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	return r.Row.Scan(dest...)
}
