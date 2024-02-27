package squirrel

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/lann/builder"
)

type dropStmt struct {
	PlaceholderFormat PlaceholderFormat
	RunWith           BaseRunner
	Prefixes          []Sqlizer
	StatementKeyword  string
	Table             string
	Suffixes          []Sqlizer
}

func (d *dropStmt) Exec() (sql.Result, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return ExecWith(d.RunWith, d)
}

func (d *dropStmt) Query() (*sql.Rows, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return QueryWith(d.RunWith, d)
}

func (d *dropStmt) ToSql() (sqlStr string, args []interface{}, err error) {
	if len(d.Table) == 0 {
		err = errors.New("drop statements must specify a table")
		return
	}

	sql := &bytes.Buffer{}

	if len(d.Prefixes) > 0 {
		args, err = appendToSql(d.Prefixes, sql, " ", args)
		if err != nil {
			return
		}

		sql.WriteString(" ")
	}

	if d.StatementKeyword == "" {
		sql.WriteString("DROP ")
	} else {
		sql.WriteString(d.StatementKeyword)
		sql.WriteString(" ")
	}

	sql.WriteString("TABLE ")
	sql.WriteString(d.Table)

	if err != nil {
		return
	}

	if len(d.Suffixes) > 0 {
		sql.WriteString(" ")
		args, err = appendToSql(d.Suffixes, sql, " ", args)
		if err != nil {
			return
		}
	}
	sqlStr, err = d.PlaceholderFormat.ReplacePlaceholders(sql.String())
	return
}

// Builder

// DropBuilder builds SQL DROP statements.
type DropBuilder builder.Builder

func init() {
	builder.Register(DropBuilder{}, dropStmt{})
}

// Format methods

// PlaceholderFormat sets PlaceholderFormat (e.g. Question or Dollar) for the
// query.
func (b DropBuilder) PlaceholderFormat(f PlaceholderFormat) DropBuilder {
	return builder.Set(b, "PlaceholderFormat", f).(DropBuilder)
}

// Runner methods

// RunWith sets a Runner (like database/sql.DB) to be used with e.g. Exec.
func (b DropBuilder) RunWith(runner BaseRunner) DropBuilder {
	return setRunWith(b, runner).(DropBuilder)
}

// Exec builds and Execs the query with the Runner set by RunWith.
func (b DropBuilder) Exec() (sql.Result, error) {
	data := builder.GetStruct(b).(dropStmt)
	return data.Exec()
}

// Query builds and Querys the query with the Runner set by RunWith.
func (b DropBuilder) Query() (*sql.Rows, error) {
	data := builder.GetStruct(b).(dropStmt)
	return data.Query()
}

// SQL methods

// ToSql builds the query into a SQL string and bound args.
func (b DropBuilder) ToSql() (string, []interface{}, error) {
	data := builder.GetStruct(b).(dropStmt)
	return data.ToSql()
}

// MustSql builds the query into a SQL string and bound args.
// It panics if there are any errors.
func (b DropBuilder) MustSql() (string, []interface{}) {
	sql, args, err := b.ToSql()
	if err != nil {
		panic(err)
	}
	return sql, args
}

// Prefix adds an expression to the beginning of the query
func (b DropBuilder) Prefix(sql string, args ...interface{}) DropBuilder {
	return b.PrefixExpr(Expr(sql, args...))
}

// PrefixExpr adds an expression to the very beginning of the query
func (b DropBuilder) PrefixExpr(expr Sqlizer) DropBuilder {
	return builder.Append(b, "Prefixes", expr).(DropBuilder)
}

// Table sets the TABLE clause of the query.
func (b DropBuilder) Table(table string) DropBuilder {
	return builder.Set(b, "Table", table).(DropBuilder)
}

// Suffix adds an expression to the end of the query
func (b DropBuilder) Suffix(sql string, args ...interface{}) DropBuilder {
	return b.SuffixExpr(Expr(sql, args...))
}

// SuffixExpr adds an expression to the end of the query
func (b DropBuilder) SuffixExpr(expr Sqlizer) DropBuilder {
	return builder.Append(b, "Suffixes", expr).(DropBuilder)
}
