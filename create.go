package squirrel

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/lann/builder"
)

type createStmt struct {
	PlaceholderFormat PlaceholderFormat
	RunWith           BaseRunner
	Prefixes          []Sqlizer
	StatementKeyword  string
	Table             string
	Columns           []string
	Types             []string
	PrimaryKey        []Sqlizer
	Suffixes          []Sqlizer
}

func (d *createStmt) Exec() (sql.Result, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return ExecWith(d.RunWith, d)
}

func (d *createStmt) Query() (*sql.Rows, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return QueryWith(d.RunWith, d)
}

func (d *createStmt) ToSql() (sqlStr string, args []interface{}, err error) {
	if len(d.Table) == 0 {
		err = errors.New("create statements must specify a table")
		return
	}
	if len(d.Columns) == 0 {
		err = errors.New("create statements must have at least one set of columns")
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
		sql.WriteString("CREATE ")
	} else {
		sql.WriteString(d.StatementKeyword)
		sql.WriteString(" ")
	}

	sql.WriteString("TABLE ")
	sql.WriteString(d.Table)
	sql.WriteString(" ")

	sql.WriteString("( ")
	err = d.appendValuesToSQL(sql)
	if len(d.PrimaryKey) > 0 {
		sql.WriteString(", PRIMARY KEY ")
		sql.WriteString("(")
		args, err = appendToSql(d.PrimaryKey, sql, ", ", args)
		sql.WriteString(") ")
		if err != nil {
			return
		}
	}
	sql.WriteString(") ")

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

func (d *createStmt) appendValuesToSQL(w io.Writer) error {
	if len(d.Columns) == 0 {
		return errors.New("columns for create statements are not set")
	}
	if len(d.Types) == 0 {
		return errors.New("types for create statements are not set")
	}
	if len(d.Types) != len(d.Columns) {
		return errors.New("types size are not equal to columns for create statements are not set")
	}
	valueStrings := make([]string, len(d.Columns))
	for i, col := range d.Columns {
		valueStrings[i] = fmt.Sprintf("%s %s", col, d.Types[i])
	}
	io.WriteString(w, strings.Join(valueStrings, ", "))

	return nil
}

// Builder

// CreateBuilder builds SQL CREATE statements.
type CreateBuilder builder.Builder

func init() {
	builder.Register(CreateBuilder{}, createStmt{})
}

// Format methods

// PlaceholderFormat sets PlaceholderFormat (e.g. Question or Dollar) for the
// query.
func (b CreateBuilder) PlaceholderFormat(f PlaceholderFormat) CreateBuilder {
	return builder.Set(b, "PlaceholderFormat", f).(CreateBuilder)
}

// Runner methods

// RunWith sets a Runner (like database/sql.DB) to be used with e.g. Exec.
func (b CreateBuilder) RunWith(runner BaseRunner) CreateBuilder {
	return setRunWith(b, runner).(CreateBuilder)
}

// Exec builds and Execs the query with the Runner set by RunWith.
func (b CreateBuilder) Exec() (sql.Result, error) {
	data := builder.GetStruct(b).(createStmt)
	return data.Exec()
}

// Query builds and Querys the query with the Runner set by RunWith.
func (b CreateBuilder) Query() (*sql.Rows, error) {
	data := builder.GetStruct(b).(createStmt)
	return data.Query()
}

// SQL methods

// ToSql builds the query into a SQL string and bound args.
func (b CreateBuilder) ToSql() (string, []interface{}, error) {
	data := builder.GetStruct(b).(createStmt)
	return data.ToSql()
}

// MustSql builds the query into a SQL string and bound args.
// It panics if there are any errors.
func (b CreateBuilder) MustSql() (string, []interface{}) {
	sql, args, err := b.ToSql()
	if err != nil {
		panic(err)
	}
	return sql, args
}

// Prefix adds an expression to the beginning of the query
func (b CreateBuilder) Prefix(sql string, args ...interface{}) CreateBuilder {
	return b.PrefixExpr(Expr(sql, args...))
}

// PrefixExpr adds an expression to the very beginning of the query
func (b CreateBuilder) PrefixExpr(expr Sqlizer) CreateBuilder {
	return builder.Append(b, "Prefixes", expr).(CreateBuilder)
}

// Table sets the TABLE clause of the query.
func (b CreateBuilder) Table(table string) CreateBuilder {
	return builder.Set(b, "Table", table).(CreateBuilder)
}

// Columns adds create columns to the query.
func (b CreateBuilder) Columns(columns ...string) CreateBuilder {
	return builder.Extend(b, "Columns", columns).(CreateBuilder)
}

// Types adds create types to the query.
func (b CreateBuilder) Types(types ...string) CreateBuilder {
	return builder.Extend(b, "Types", types).(CreateBuilder)
}

// PrimaryKey adds an primary key to the query
func (b CreateBuilder) PrimaryKey(sql string, args ...interface{}) CreateBuilder {
	return b.PrimaryKeyExpr(Expr(sql, args...))
}

// PrimaryKeyExpr adds an expression to the query
func (b CreateBuilder) PrimaryKeyExpr(expr Sqlizer) CreateBuilder {
	return builder.Append(b, "PrimaryKey", expr).(CreateBuilder)
}

// Suffix adds an expression to the end of the query
func (b CreateBuilder) Suffix(sql string, args ...interface{}) CreateBuilder {
	return b.SuffixExpr(Expr(sql, args...))
}

// SuffixExpr adds an expression to the end of the query
func (b CreateBuilder) SuffixExpr(expr Sqlizer) CreateBuilder {
	return builder.Append(b, "Suffixes", expr).(CreateBuilder)
}

// SetMap set columns and values for insert builder from a map of column name and value
// note that it will reset all previous columns and values was set if any
func (b CreateBuilder) SetMap(clauses map[string]string) CreateBuilder {
	// Keep the columns in a consistent order by sorting the column key string.
	cols := make([]string, 0, len(clauses))
	for col := range clauses {
		cols = append(cols, col+" "+clauses[col])
	}
	sort.Strings(cols)

	vals := make([]string, 0, len(clauses))
	for _, col := range cols {
		vals = append(vals, clauses[col])
	}

	b = builder.Set(b, "Columns", cols).(CreateBuilder)
	b = builder.Set(b, "Types", vals).(CreateBuilder)

	return b
}
