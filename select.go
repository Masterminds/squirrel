package squirrel

import (
	"database/sql"
	"fmt"

	"github.com/lann/builder"
)

type selectData struct {
	PlaceholderFormat PlaceholderFormat
	RunWith           BaseRunner
	SerializeWith     Serializer
	Prefixes          exprs
	Options           []string
	Columns           []Sqlizer
	From              Sqlizer
	Joins             []Sqlizer
	WhereParts        []Sqlizer
	GroupBys          []string
	HavingParts       []Sqlizer
	OrderBys          []string
	Limit             string
	Offset            string
	Suffixes          exprs
}

func (d *selectData) Exec() (sql.Result, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	if d.SerializeWith == nil {
		return nil, SerializerNotSet
	}
	return ExecWith(d.RunWith, d, d.SerializeWith)
}

func (d *selectData) Query() (*sql.Rows, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	if d.SerializeWith == nil {
		return nil, SerializerNotSet
	}
	return QueryWith(d.RunWith, d, d.SerializeWith)
}

func (d *selectData) QueryRow() RowScanner {
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

func (d *selectData) ToSql() (sqlStr string, args []interface{}, err error) {
	return d.ToSqlWithSerializer(DefaultSerializer{})
}

func (d *selectData) ToSqlWithSerializer(serializer Serializer) (sqlStr string, args []interface{}, err error) {
	if len(d.Columns) == 0 {
		err = fmt.Errorf("select statements must have at least one result column")
		return
	}

	return serializer.Select(*d)
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

// SerializeWith sets a Serializer (that is, db specific writer) to be used with.
func (b SelectBuilder) SerializeWith(serializer Serializer) SelectBuilder {
	return setSerializeWith(b, serializer).(SelectBuilder)
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

// ToSql builds the query into a SQL string and bound args with the default serializer.
func (b SelectBuilder) ToSql() (sqlStr string, args []interface{}, err error) {
	return b.ToSqlWithSerializer(DefaultSerializer{})
}

// ToSql builds the query into a SQL string and bound args with a specific serializer.
func (b SelectBuilder) ToSqlWithSerializer(serializer Serializer) (string, []interface{}, error) {
	data := builder.GetStruct(b).(selectData)
	return data.ToSqlWithSerializer(serializer)
}

// Prefix adds an expression to the beginning of the query
func (b SelectBuilder) Prefix(sql string, args ...interface{}) SelectBuilder {
	return builder.Append(b, "Prefixes", Expr(sql, args...)).(SelectBuilder)
}

// Distinct adds a DISTINCT clause to the query.
func (b SelectBuilder) Distinct() SelectBuilder {
	return b.Options("DISTINCT")
}

// Options adds select option to the query
func (b SelectBuilder) Options(options ...string) SelectBuilder {
	return builder.Extend(b, "Options", options).(SelectBuilder)
}

// Columns adds result columns to the query.
func (b SelectBuilder) Columns(columns ...string) SelectBuilder {
	var parts []interface{}
	for _, str := range columns {
		parts = append(parts, newPart(str))
	}
	return builder.Extend(b, "Columns", parts).(SelectBuilder)
}

// Column adds a result column to the query.
// Unlike Columns, Column accepts args which will be bound to placeholders in
// the columns string, for example:
//   Column("IF(col IN ("+squirrel.Placeholders(3)+"), 1, 0) as col", 1, 2, 3)
func (b SelectBuilder) Column(column interface{}, args ...interface{}) SelectBuilder {
	return builder.Append(b, "Columns", newPart(column, args...)).(SelectBuilder)
}

// From sets the FROM clause of the query.
func (b SelectBuilder) From(from string) SelectBuilder {
	return builder.Set(b, "From", newPart(from)).(SelectBuilder)
}

// FromSelect sets a subquery into the FROM clause of the query.
func (b SelectBuilder) FromSelect(from SelectBuilder, alias string) SelectBuilder {
	return builder.Set(b, "From", Alias(from, alias)).(SelectBuilder)
}

// JoinClause adds a join clause to the query.
func (b SelectBuilder) JoinClause(pred interface{}, args ...interface{}) SelectBuilder {
	return builder.Append(b, "Joins", newPart(pred, args...)).(SelectBuilder)
}

// Join adds a JOIN clause to the query.
func (b SelectBuilder) Join(join string, rest ...interface{}) SelectBuilder {
	return b.JoinClause("JOIN "+join, rest...)
}

// LeftJoin adds a LEFT JOIN clause to the query.
func (b SelectBuilder) LeftJoin(join string, rest ...interface{}) SelectBuilder {
	return b.JoinClause("LEFT JOIN "+join, rest...)
}

// RightJoin adds a RIGHT JOIN clause to the query.
func (b SelectBuilder) RightJoin(join string, rest ...interface{}) SelectBuilder {
	return b.JoinClause("RIGHT JOIN "+join, rest...)
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
