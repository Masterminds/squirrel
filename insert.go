package squirrel

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/lann/builder"
)

type insertData struct {
	PlaceholderFormat PlaceholderFormat
	RunWith           BaseRunner
	SerializeWith     Serializer
	Prefixes          exprs
	Options           []string
	Into              string
	Columns           []string
	Values            [][]interface{}
	Suffixes          exprs
	Select            *SelectBuilder
}

func (d *insertData) Exec() (sql.Result, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	if d.SerializeWith == nil {
		return nil, SerializerNotSet
	}
	return ExecWith(d.RunWith, d, d.SerializeWith)
}

func (d *insertData) Query() (*sql.Rows, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	if d.SerializeWith == nil {
		return nil, SerializerNotSet
	}
	return QueryWith(d.RunWith, d, d.SerializeWith)
}

func (d *insertData) QueryRow() RowScanner {
	if d.RunWith == nil {
		return &Row{err: RunnerNotSet}
	}
	if d.SerializeWith == nil {
		return &Row{err: SerializerNotSet}
	}
	queryRower, ok := d.RunWith.(QueryRower)
	if !ok {
		return &Row{err: RunnerNotQueryRunner}
	}
	return QueryRowWith(queryRower, d, d.SerializeWith)
}

func (d *insertData) ToSql() (sqlStr string, args []interface{}, err error) {
	return d.ToSqlWithSerializer(DefaultSerializer{})
}

func (d *insertData) ToSqlWithSerializer(serializer Serializer) (sqlStr string, args []interface{}, err error) {
	if len(d.Into) == 0 {
		err = errors.New("insert statements must specify a table")
		return
	}
	if len(d.Values) == 0 && d.Select == nil {
		err = errors.New("insert statements must have at least one set of values or select clause")
		return
	}

	return serializer.Insert(*d)
}

func (d *insertData) appendValuesToSQL(w io.Writer, args []interface{}) ([]interface{}, error) {
	if len(d.Values) == 0 {
		return args, errors.New("values for insert statements are not set")
	}

	io.WriteString(w, "VALUES ")

	valuesStrings := make([]string, len(d.Values))
	for r, row := range d.Values {
		valueStrings := make([]string, len(row))
		for v, val := range row {
			e, isExpr := val.(expr)
			if isExpr {
				valueStrings[v] = e.sql
				args = append(args, e.args...)
			} else {
				valueStrings[v] = "?"
				args = append(args, val)
			}
		}
		valuesStrings[r] = fmt.Sprintf("(%s)", strings.Join(valueStrings, ","))
	}

	io.WriteString(w, strings.Join(valuesStrings, ","))

	return args, nil
}

func (d *insertData) appendSelectToSQL(w io.Writer, args []interface{}, serializer Serializer) ([]interface{}, error) {
	if d.Select == nil {
		return args, errors.New("select clause for insert statements are not set")
	}

	selectClause, sArgs, err := d.Select.ToSqlWithSerializer(serializer)
	if err != nil {
		return args, err
	}

	io.WriteString(w, selectClause)
	args = append(args, sArgs...)

	return args, nil
}

// Builder

// InsertBuilder builds SQL INSERT statements.
type InsertBuilder builder.Builder

func init() {
	builder.Register(InsertBuilder{}, insertData{})
}

// Format methods

// PlaceholderFormat sets PlaceholderFormat (e.g. Question or Dollar) for the
// query.
func (b InsertBuilder) PlaceholderFormat(f PlaceholderFormat) InsertBuilder {
	return builder.Set(b, "PlaceholderFormat", f).(InsertBuilder)
}

// Runner methods

// RunWith sets a Runner (like database/sql.DB) to be used with e.g. Exec.
func (b InsertBuilder) RunWith(runner BaseRunner) InsertBuilder {
	return setRunWith(b, runner).(InsertBuilder)
}

// SerializeWith sets a Serializer (that is, db specific writer) to be used with.
func (b InsertBuilder) SerializeWith(serializer Serializer) InsertBuilder {
	return setSerializeWith(b, serializer).(InsertBuilder)
}

// Exec builds and Execs the query with the Runner set by RunWith.
func (b InsertBuilder) Exec() (sql.Result, error) {
	data := builder.GetStruct(b).(insertData)
	return data.Exec()
}

// Query builds and Querys the query with the Runner set by RunWith.
func (b InsertBuilder) Query() (*sql.Rows, error) {
	data := builder.GetStruct(b).(insertData)
	return data.Query()
}

// QueryRow builds and QueryRows the query with the Runner set by RunWith.
func (b InsertBuilder) QueryRow() RowScanner {
	data := builder.GetStruct(b).(insertData)
	return data.QueryRow()
}

// Scan is a shortcut for QueryRow().Scan.
func (b InsertBuilder) Scan(dest ...interface{}) error {
	return b.QueryRow().Scan(dest...)
}

// SQL methods

// ToSql builds the query into a SQL string and bound args with the default serializer.
func (b InsertBuilder) ToSql() (sqlStr string, args []interface{}, err error) {
	return b.ToSqlWithSerializer(DefaultSerializer{})
}

// ToSql builds the query into a SQL string and bound args with a specific serializer.
func (b InsertBuilder) ToSqlWithSerializer(serializer Serializer) (string, []interface{}, error) {
	data := builder.GetStruct(b).(insertData)
	return data.ToSqlWithSerializer(serializer)
}

// Prefix adds an expression to the beginning of the query
func (b InsertBuilder) Prefix(sql string, args ...interface{}) InsertBuilder {
	return builder.Append(b, "Prefixes", Expr(sql, args...)).(InsertBuilder)
}

// Options adds keyword options before the INTO clause of the query.
func (b InsertBuilder) Options(options ...string) InsertBuilder {
	return builder.Extend(b, "Options", options).(InsertBuilder)
}

// Into sets the INTO clause of the query.
func (b InsertBuilder) Into(from string) InsertBuilder {
	return builder.Set(b, "Into", from).(InsertBuilder)
}

// Columns adds insert columns to the query.
func (b InsertBuilder) Columns(columns ...string) InsertBuilder {
	return builder.Extend(b, "Columns", columns).(InsertBuilder)
}

// Values adds a single row's values to the query.
func (b InsertBuilder) Values(values ...interface{}) InsertBuilder {
	return builder.Append(b, "Values", values).(InsertBuilder)
}

// Suffix adds an expression to the end of the query
func (b InsertBuilder) Suffix(sql string, args ...interface{}) InsertBuilder {
	return builder.Append(b, "Suffixes", Expr(sql, args...)).(InsertBuilder)
}

// SetMap set columns and values for insert builder from a map of column name and value
// note that it will reset all previous columns and values was set if any
func (b InsertBuilder) SetMap(clauses map[string]interface{}) InsertBuilder {
	cols := make([]string, 0, len(clauses))
	vals := make([]interface{}, 0, len(clauses))
	for col, val := range clauses {
		cols = append(cols, col)
		vals = append(vals, val)
	}

	b = builder.Set(b, "Columns", cols).(InsertBuilder)
	b = builder.Set(b, "Values", [][]interface{}{vals}).(InsertBuilder)
	return b
}

// Select set Select clause for insert query
// If Values and Select are used, then Select has higher priority
func (b InsertBuilder) Select(sb SelectBuilder) InsertBuilder {
	return builder.Set(b, "Select", &sb).(InsertBuilder)
}
