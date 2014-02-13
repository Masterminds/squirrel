package squirrel

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/lann/builder"
	"strings"
)

type insertData struct {
	PlaceholderFormat PlaceholderFormat
	RunWith           Runner
	Into              string
	Columns           []string
	Values            [][]interface{}
}

func (d *insertData) Exec() (sql.Result, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return ExecWith(d.RunWith, d)
}

func (d *insertData) ToSql() (sqlStr string, args []interface{}, err error) {
	if len(d.Into) == 0 {
		err = fmt.Errorf("insert statements must specify a table")
		return
	}
	if len(d.Values) == 0 {
		err = fmt.Errorf("insert statements must have at least one set of values")
		return
	}

	var sql bytes.Buffer

	sql.WriteString("INSERT INTO ")
	sql.WriteString(d.Into)
	sql.WriteString(" ")

	if len(d.Columns) > 0 {
		sql.WriteString("(")
		sql.WriteString(strings.Join(d.Columns, ","))
		sql.WriteString(") ")
	}

	sql.WriteString("VALUES ")

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
	sql.WriteString(strings.Join(valuesStrings, ","))

	sqlStr, err = d.PlaceholderFormat.ReplacePlaceholders(sql.String())
	return
}

type expr struct {
	sql string
	args []interface{}
}

// Expr builds value expressions for InsertBuilder.Values.
//
// Ex:
//     .Values(Expr("FROM_UNIXTIME(?)", t))
func Expr(sql string, args ...interface{}) expr {
	return expr{sql: sql, args: args}
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
func (b InsertBuilder) RunWith(runner Runner) InsertBuilder {
	return builder.Set(b, "RunWith", runner).(InsertBuilder)
}

// Exec builds and Execs the query with the Runner set by RunWith.
func (b InsertBuilder) Exec() (sql.Result, error) {
	data := builder.GetStruct(b).(insertData)
	return data.Exec()
}

// SQL methods

// ToSql builds the query into a SQL string and bound args.
func (b InsertBuilder) ToSql() (string, []interface{}, error) {
	data := builder.GetStruct(b).(insertData)
	return data.ToSql()
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
