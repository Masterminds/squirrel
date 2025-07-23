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

type Typed struct {
	Type  string
	Value interface{}
}

type mergeData struct {
	PlaceholderFormat PlaceholderFormat
	RunWith           BaseRunner
	Prefixes          []Sqlizer
	Into              string
	ValuesAlias       string
	On                string
	When              []string
	Columns           []string
	Values            [][]interface{}
	Suffixes          []Sqlizer
	Select            *SelectBuilder
}

func (d *mergeData) Exec() (sql.Result, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return ExecWith(d.RunWith, d)
}

func (d *mergeData) Query() (*sql.Rows, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return QueryWith(d.RunWith, d)
}

func (d *mergeData) QueryRow() RowScanner {
	if d.RunWith == nil {
		return &Row{err: RunnerNotSet}
	}
	queryRower, ok := d.RunWith.(QueryRower)
	if !ok {
		return &Row{err: RunnerNotQueryRunner}
	}
	return QueryRowWith(queryRower, d)
}

func (d *mergeData) ToSql() (sqlStr string, args []interface{}, err error) {
	if len(d.Into) == 0 {
		err = ErrNoTable
		return
	}
	if len(d.Values) == 0 && d.Select == nil {
		err = ErrNoValues
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

	sql.WriteString("MERGE INTO ")
	sql.WriteString(d.Into)
	sql.WriteString(" ")

	sql.WriteString("USING ")

	sql.WriteString("(")
	if d.Select != nil {
		args, err = d.appendSelectToSQL(sql, args)
	} else {
		args, err = d.appendValuesToSQL(sql, args)
	}
	if err != nil {
		return
	}

	sql.WriteString(")")

	if d.ValuesAlias != "" {
		sql.WriteString(" AS ")
		sql.WriteString(d.ValuesAlias)
		sql.WriteString(" ")
	}

	if len(d.Columns) > 0 {
		sql.WriteString("(")
		sql.WriteString(strings.Join(d.Columns, ","))
		sql.WriteString(")")
	}

	if d.On != "" {
		sql.WriteString(" ON ")
		sql.WriteString(d.On)
	}

	if len(d.When) > 0 {
		sql.WriteString(" WHEN ")
		sql.WriteString(strings.Join(d.When, " WHEN "))
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

func (d *mergeData) appendValuesToSQL(w io.Writer, args []interface{}) ([]interface{}, error) {
	if len(d.Values) == 0 {
		return args, errors.New("values for insert statements are not set")
	}

	io.WriteString(w, "VALUES ")

	valuesStrings := make([]string, len(d.Values))
	for r, row := range d.Values {
		valueStrings := make([]string, len(row))
		for v, val := range row {
			var valueType string
			switch rv := val.(type) {
			case Typed:
				valueType = rv.Type
				val = rv.Value
			}
			if vs, ok := val.(Sqlizer); ok {
				vsql, vargs, err := vs.ToSql()
				if err != nil {
					return nil, err
				}
				valueStrings[v] = vsql
				args = append(args, vargs...)
			} else {
				valueStrings[v] = "?"
				args = append(args, val)
			}
			if valueType != "" {
				valueStrings[v] = fmt.Sprintf("%s::%s", valueStrings[v], valueType)
			}
		}
		valuesStrings[r] = fmt.Sprintf("(%s)", strings.Join(valueStrings, ","))
	}

	io.WriteString(w, strings.Join(valuesStrings, ","))

	return args, nil
}

func (d *mergeData) appendSelectToSQL(w io.Writer, args []interface{}) ([]interface{}, error) {
	if d.Select == nil {
		return args, errors.New("select clause for insert statements are not set")
	}

	selectClause, sArgs, err := d.Select.ToSql()
	if err != nil {
		return args, err
	}

	io.WriteString(w, selectClause)
	args = append(args, sArgs...)

	return args, nil
}

// Builder

// MergeBuilder builds SQL INSERT statements.
type MergeBuilder builder.Builder

func init() {
	builder.Register(MergeBuilder{}, mergeData{})
}

// Format methods

// PlaceholderFormat sets PlaceholderFormat (e.g. Question or Dollar) for the
// query.
func (b MergeBuilder) PlaceholderFormat(f PlaceholderFormat) MergeBuilder {
	return builder.Set(b, "PlaceholderFormat", f).(MergeBuilder)
}

// Runner methods

// RunWith sets a Runner (like database/sql.DB) to be used with e.g. Exec.
func (b MergeBuilder) RunWith(runner BaseRunner) MergeBuilder {
	return setRunWith(b, runner).(MergeBuilder)
}

// Exec builds and Execs the query with the Runner set by RunWith.
func (b MergeBuilder) Exec() (sql.Result, error) {
	data := builder.GetStruct(b).(mergeData)
	return data.Exec()
}

// Query builds and Querys the query with the Runner set by RunWith.
func (b MergeBuilder) Query() (*sql.Rows, error) {
	data := builder.GetStruct(b).(mergeData)
	return data.Query()
}

// QueryRow builds and QueryRows the query with the Runner set by RunWith.
func (b MergeBuilder) QueryRow() RowScanner {
	data := builder.GetStruct(b).(mergeData)
	return data.QueryRow()
}

// Scan is a shortcut for QueryRow().Scan.
func (b MergeBuilder) Scan(dest ...interface{}) error {
	return b.QueryRow().Scan(dest...)
}

// SQL methods

// ToSql builds the query into a SQL string and bound args.
func (b MergeBuilder) ToSql() (string, []interface{}, error) {
	data := builder.GetStruct(b).(mergeData)
	return data.ToSql()
}

// MustSql builds the query into a SQL string and bound args.
// It panics if there are any errors.
func (b MergeBuilder) MustSql() (string, []interface{}) {
	sql, args, err := b.ToSql()
	if err != nil {
		panic(err)
	}
	return sql, args
}

// Prefix adds an expression to the beginning of the query
func (b MergeBuilder) Prefix(sql string, args ...interface{}) MergeBuilder {
	return b.PrefixExpr(Expr(sql, args...))
}

// PrefixExpr adds an expression to the very beginning of the query
func (b MergeBuilder) PrefixExpr(expr Sqlizer) MergeBuilder {
	return builder.Append(b, "Prefixes", expr).(MergeBuilder)
}

// Into sets the INTO clause of the query.
func (b MergeBuilder) Into(into string) MergeBuilder {
	return builder.Set(b, "Into", into).(MergeBuilder)
}

// ValuesAlias sets the AS vals clause of the query.
func (b MergeBuilder) ValuesAlias(valuesAlias string) MergeBuilder {
	return builder.Set(b, "ValuesAlias", valuesAlias).(MergeBuilder)
}

// On sets the ON clause of the query.
func (b MergeBuilder) On(on string) MergeBuilder {
	return builder.Set(b, "On", on).(MergeBuilder)
}

// When sets the WHEN MATCHED/NOT MATCHED clause of the query.
func (b MergeBuilder) When(when string) MergeBuilder {
	return builder.Append(b, "When", when).(MergeBuilder)
}

// Columns adds insert columns to the query.
func (b MergeBuilder) Columns(columns ...string) MergeBuilder {
	return builder.Extend(b, "Columns", columns).(MergeBuilder)
}

// Values adds a single row's values to the query.
func (b MergeBuilder) Values(values ...interface{}) MergeBuilder {
	return builder.Append(b, "Values", values).(MergeBuilder)
}

// Suffix adds an expression to the end of the query
func (b MergeBuilder) Suffix(sql string, args ...interface{}) MergeBuilder {
	return b.SuffixExpr(Expr(sql, args...))
}

// SuffixExpr adds an expression to the end of the query
func (b MergeBuilder) SuffixExpr(expr Sqlizer) MergeBuilder {
	return builder.Append(b, "Suffixes", expr).(MergeBuilder)
}

// SetMap set columns and values for insert builder from a map of column name and value
// note that it will reset all previous columns and values was set if any
func (b MergeBuilder) SetMap(clauses map[string]interface{}) MergeBuilder {
	// Keep the columns in a consistent order by sorting the column key string.
	cols := make([]string, 0, len(clauses))
	for col := range clauses {
		cols = append(cols, col)
	}
	sort.Strings(cols)

	vals := make([]interface{}, 0, len(clauses))
	for _, col := range cols {
		vals = append(vals, clauses[col])
	}

	b = builder.Set(b, "Columns", cols).(MergeBuilder)
	b = builder.Set(b, "Values", [][]interface{}{vals}).(MergeBuilder)

	return b
}

// Select set Select clause for insert query
// If Values and Select are used, then Select has higher priority
func (b MergeBuilder) Select(sb SelectBuilder) MergeBuilder {
	return builder.Set(b, "Select", &sb).(MergeBuilder)
}
