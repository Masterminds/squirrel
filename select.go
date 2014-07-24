package squirrel

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/lann/builder"
	"strings"
	"io"
)

type selectData struct {
	PlaceholderFormat PlaceholderFormat
	RunWith           Runner
	Prefixes          exprs
	Distinct          bool
	Columns           selectParts
	From              string
	Joins             []string
	WhereParts        whereParts
	GroupBys          []string
	HavingParts       whereParts
	OrderBys          []string
	Limit             string
	Offset            string
	Suffixes          exprs
}

type selectPart struct {
	part
}

type selectParts []sqlSerializable

func newSelectPart(pred interface{}, args ...interface{}) selectPart {
	return selectPart{part:part{pred: pred, args: args}}
}

func (parts selectParts) AppendToSql(w io.Writer, sep string, args []interface{}) ([]interface{}, error) {
	return appendToSql(parts, w, sep, args)
}

func (d *selectData) Exec() (sql.Result, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return ExecWith(d.RunWith, d)
}

func (d *selectData) Query() (*sql.Rows, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	return QueryWith(d.RunWith, d)
}

func (d *selectData) QueryRow() RowScanner {
	if d.RunWith == nil {
		return &Row{err: RunnerNotSet}
	}
	return QueryRowWith(d.RunWith, d)
}

func (d *selectData) ToSql() (sqlStr string, args []interface{}, err error) {
	if len(d.Columns) == 0 {
		err = fmt.Errorf("select statements must have at least one result column")
		return
	}

	sql := &bytes.Buffer{}

	if len(d.Prefixes) > 0 {
		args, _ = d.Prefixes.AppendToSql(sql, " ", args)
		sql.WriteString(" ")
	}

	sql.WriteString("SELECT ")

	if d.Distinct {
		sql.WriteString("DISTINCT ")
	}

	if len(d.Columns) > 0 {
		args, err = d.Columns.AppendToSql(sql, ", ", args)
		if err != nil {
			return
		}
	}

	if len(d.From) > 0 {
		sql.WriteString(" FROM ")
		sql.WriteString(d.From)
	}

	if len(d.Joins) > 0 {
		sql.WriteString(" ")
		sql.WriteString(strings.Join(d.Joins, " "))
	}

	if len(d.WhereParts) > 0 {
		sql.WriteString(" WHERE ")
		args, err = d.WhereParts.AppendToSql(sql, " AND ", args)
		if err != nil {
			return
		}
	}

	if len(d.GroupBys) > 0 {
		sql.WriteString(" GROUP BY ")
		sql.WriteString(strings.Join(d.GroupBys, ", "))
	}

	if len(d.HavingParts) > 0 {
		sql.WriteString(" HAVING ")
		args, err = d.HavingParts.AppendToSql(sql, " AND ", args)
		if err != nil {
			return
		}
	}

	if len(d.OrderBys) > 0 {
		sql.WriteString(" ORDER BY ")
		sql.WriteString(strings.Join(d.OrderBys, ", "))
	}

	if len(d.Limit) > 0 {
		sql.WriteString(" LIMIT ")
		sql.WriteString(d.Limit)
	}

	if len(d.Offset) > 0 {
		sql.WriteString(" OFFSET ")
		sql.WriteString(d.Offset)
	}

	if len(d.Suffixes) > 0 {
		sql.WriteString(" ")
		args, _ = d.Suffixes.AppendToSql(sql, " ", args)
	}

	sqlStr, err = d.PlaceholderFormat.ReplacePlaceholders(sql.String())
	return
}

// Builder

// SelectBuilder builds SQL SELECT statements.
type SelectBuilder builder.Builder

func init() {
	builder.Register(SelectBuilder{}, selectData{})
}

// Format methods

// PlaceholderFormat sets PlaceholderFormat (e.g. Question or Dollar) for the
// query.
func (b SelectBuilder) PlaceholderFormat(f PlaceholderFormat) SelectBuilder {
	return builder.Set(b, "PlaceholderFormat", f).(SelectBuilder)
}

// Runner methods

// RunWith sets a Runner (like database/sql.DB) to be used with e.g. Exec.
func (b SelectBuilder) RunWith(runner BaseRunner) SelectBuilder {
	return setRunWith(b, runner).(SelectBuilder)
}

// Exec builds and Execs the query with the Runner set by RunWith.
func (b SelectBuilder) Exec() (sql.Result, error) {
	data := builder.GetStruct(b).(selectData)
	return data.Exec()
}

// Query builds and Querys the query with the Runner set by RunWith.
func (b SelectBuilder) Query() (*sql.Rows, error) {
	data := builder.GetStruct(b).(selectData)
	return data.Query()
}

// QueryRow builds and QueryRows the query with the Runner set by RunWith.
func (b SelectBuilder) QueryRow() RowScanner {
	data := builder.GetStruct(b).(selectData)
	return data.QueryRow()
}

// Scan is a shortcut for QueryRow().Scan.
func (b SelectBuilder) Scan(dest ...interface{}) error {
	return b.QueryRow().Scan(dest...)
}

// SQL methods

// ToSql builds the query into a SQL string and bound args.
func (b SelectBuilder) ToSql() (string, []interface{}, error) {
	data := builder.GetStruct(b).(selectData)
	return data.ToSql()
}

// Prefix adds an expression to the beginning of the query
func (b SelectBuilder) Prefix(sql string, args ...interface{}) SelectBuilder {
	return builder.Append(b, "Prefixes", Expr(sql, args...)).(SelectBuilder)
}

// Distinct adds a DISTINCT clause to the query.
func (b SelectBuilder) Distinct() SelectBuilder {
	return builder.Set(b, "Distinct", true).(SelectBuilder)
}

// Columns adds result columns to the query.
func (b SelectBuilder) Columns(columns ...string) SelectBuilder {
	var parts []interface{}
	for _, str := range columns {
		parts = append(parts, newSelectPart(str))
	}
	return builder.Extend(b, "Columns", parts).(SelectBuilder)
}

// Column add result column to query with arguments
// use it like Column("IF(col IN ("+squirel.Placeholders(3)+"), 1, 0) as col", 1, 2, 3)
// or with interface slice, see https://code.google.com/p/go-wiki/wiki/InterfaceSlice
// Column("IF(col IN ("+squirel.Placeholders(len(interfaceSlice))+"), 1, 0) as col", interfaceSlice...)
func (b SelectBuilder) Column(column interface{}, args ...interface{}) SelectBuilder {
	return builder.Append(b, "Columns", newSelectPart(column, args...)).(SelectBuilder)
}

// From sets the FROM clause of the query.
func (b SelectBuilder) From(from string) SelectBuilder {
	return builder.Set(b, "From", from).(SelectBuilder)
}

// JoinClause adds a join clause to the query.
func (b SelectBuilder) JoinClause(join string) SelectBuilder {
	return builder.Append(b, "Joins", join).(SelectBuilder)
}

// Join adds a JOIN clause to the query.
func (b SelectBuilder) Join(join string) SelectBuilder {
	return b.JoinClause("JOIN " + join)
}

// LeftJoin adds a LEFT JOIN clause to the query.
func (b SelectBuilder) LeftJoin(join string) SelectBuilder {
	return b.JoinClause("LEFT JOIN " + join)
}

// RightJoin adds a RIGHT JOIN clause to the query.
func (b SelectBuilder) RightJoin(join string) SelectBuilder {
	return b.JoinClause("RIGHT JOIN " + join)
}

// Where adds an expression to the WHERE clause of the query.
//
// Expressions are ANDed together in the generated SQL.
//
// Where accepts several types for its pred argument:
//
// nil OR "" - ignored.
//
// string - SQL expression.
// If the expression has SQL placeholders then a set of arguments must be passed
// as well, one for each placeholder.
//
// map[string]interface{} OR Eq - map of SQL expressions to values. Each key is
// transformed into an expression like "<key> = ?", with the corresponding value
// bound to the placeholder. If the value is nil, the expression will be "<key>
// IS NULL". If the value is an array or slice, the expression will be "<key> IN
// (?,?,...)", with one placeholder for each item in the value. These expressions
// are ANDed together.
//
// Where will panic if pred isn't any of the above types.
func (b SelectBuilder) Where(pred interface{}, args ...interface{}) SelectBuilder {
	return builder.Append(b, "WhereParts", newWherePart(pred, args...)).(SelectBuilder)
}

// GroupBy adds GROUP BY expressions to the query.
func (b SelectBuilder) GroupBy(groupBys ...string) SelectBuilder {
	return builder.Extend(b, "GroupBys", groupBys).(SelectBuilder)
}

// Having adds an expression to the HAVING clause of the query.
//
// See Where.
func (b SelectBuilder) Having(pred interface{}, rest ...interface{}) SelectBuilder {
	return builder.Append(b, "HavingParts", newWherePart(pred, rest...)).(SelectBuilder)
}

// OrderBy adds ORDER BY expressions to the query.
func (b SelectBuilder) OrderBy(orderBys ...string) SelectBuilder {
	return builder.Extend(b, "OrderBys", orderBys).(SelectBuilder)
}

// Limit sets a LIMIT clause on the query.
func (b SelectBuilder) Limit(limit uint64) SelectBuilder {
	return builder.Set(b, "Limit", fmt.Sprintf("%d", limit)).(SelectBuilder)
}

// Offset sets a OFFSET clause on the query.
func (b SelectBuilder) Offset(offset uint64) SelectBuilder {
	return builder.Set(b, "Offset", fmt.Sprintf("%d", offset)).(SelectBuilder)
}

// Suffix adds an expression to the end of the query
func (b SelectBuilder) Suffix(sql string, args ...interface{}) SelectBuilder {
	return builder.Append(b, "Suffixes", Expr(sql, args...)).(SelectBuilder)
}
