package squirrel

import "database/sql"

type any interface{}

type sqlizer interface {
	ToSql() (string, []interface{}, error)
}

func ExecWith(db sql.DB, ts sqlizer) (res sql.Result, err error) {
	query, args, err := ts.ToSql()
	if err == nil {
		res, err = db.Exec(query, args)
	}
	return
}

func QueryWith(db sql.DB, ts sqlizer) (rows *sql.Rows, err error) {
	query, args, err := ts.ToSql()
	if err == nil {
		rows, err = db.Query(query, args)
	}
	return
}
