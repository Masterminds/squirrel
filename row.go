package squirrel

type RowScanner interface {
	Scan(...interface{}) error
}

// Proxy for database/sql.Row that lets us set our own errors
type Row struct {
	RowScanner
	err error
}

// See database/sql.Row.Scan
func (r *Row) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	return r.RowScanner.Scan(dest...)
}
