package squirrel

import (
	"database/sql"
	"fmt"

	"github.com/lann/builder"
)

type deleteData struct {
	PlaceholderFormat PlaceholderFormat
	RunWith           BaseRunner
	SerializeWith     Serializer
	Prefixes          exprs
	From              string
	WhereParts        []Sqlizer
	OrderBys          []string
	Limit             string
	Offset            string
	Suffixes          exprs
}

func (d *deleteData) Exec() (sql.Result, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	if d.SerializeWith == nil {
		return nil, SerializerNotSet
	}
	return ExecWith(d.RunWith, d, d.SerializeWith)
}

func (d *deleteData) ToSql(serializer Serializer) (sqlStr string, args []interface{}, err error) {
	if len(d.From) == 0 {
		err = fmt.Errorf("delete statements must specify a From table")
		return
	}

	return serializer.Delete(*d)
}

// Builder

// DeleteBuilder builds SQL DELETE statements.
type DeleteBuilder builder.Builder

func init() {
	builder.Register(DeleteBuilder{}, deleteData{})
}

// Format methods

// PlaceholderFormat sets PlaceholderFormat (e.g. Question or Dollar) for the
// query.
func (b DeleteBuilder) PlaceholderFormat(f PlaceholderFormat) DeleteBuilder {
	return builder.Set(b, "PlaceholderFormat", f).(DeleteBuilder)
}

// Runner methods

// RunWith sets a Runner (like database/sql.DB) to be used with e.g. Exec.
func (b DeleteBuilder) RunWith(runner BaseRunner) DeleteBuilder {
	return setRunWith(b, runner).(DeleteBuilder)
}

// SerializeWith sets a Serializer (that is, db specific writer) to be used with.
func (b DeleteBuilder) SerializeWith(serializer Serializer) DeleteBuilder {
	return setSerializeWith(b, serializer).(DeleteBuilder)
}

// Exec builds and Execs the query with the Runner set by RunWith.
func (b DeleteBuilder) Exec() (sql.Result, error) {
	data := builder.GetStruct(b).(deleteData)
	return data.Exec()
}

// SQL methods

// ToSql builds the query into a SQL string and bound args.
func (b DeleteBuilder) ToSql(serializer Serializer) (string, []interface{}, error) {
	data := builder.GetStruct(b).(deleteData)
	return data.ToSql(serializer)
}

// Prefix adds an expression to the beginning of the query
func (b DeleteBuilder) Prefix(sql string, args ...interface{}) DeleteBuilder {
	return builder.Append(b, "Prefixes", Expr(sql, args...)).(DeleteBuilder)
}

// From sets the table to be deleted from.
func (b DeleteBuilder) From(from string) DeleteBuilder {
	return builder.Set(b, "From", from).(DeleteBuilder)
}

// Where adds WHERE expressions to the query.
//
// See SelectBuilder.Where for more information.
func (b DeleteBuilder) Where(pred interface{}, args ...interface{}) DeleteBuilder {
	return builder.Append(b, "WhereParts", newWherePart(pred, args...)).(DeleteBuilder)
}

// OrderBy adds ORDER BY expressions to the query.
func (b DeleteBuilder) OrderBy(orderBys ...string) DeleteBuilder {
	return builder.Extend(b, "OrderBys", orderBys).(DeleteBuilder)
}

// Limit sets a LIMIT clause on the query.
func (b DeleteBuilder) Limit(limit uint64) DeleteBuilder {
	return builder.Set(b, "Limit", fmt.Sprintf("%d", limit)).(DeleteBuilder)
}

// Offset sets a OFFSET clause on the query.
func (b DeleteBuilder) Offset(offset uint64) DeleteBuilder {
	return builder.Set(b, "Offset", fmt.Sprintf("%d", offset)).(DeleteBuilder)
}

// Suffix adds an expression to the end of the query
func (b DeleteBuilder) Suffix(sql string, args ...interface{}) DeleteBuilder {
	return builder.Append(b, "Suffixes", Expr(sql, args...)).(DeleteBuilder)
}
