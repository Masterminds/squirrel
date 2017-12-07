package squirrel

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/lann/builder"
)

type updateData struct {
	PlaceholderFormat PlaceholderFormat
	RunWith           BaseRunner
	SerializeWith     Serializer

	Prefixes   exprs
	Table      string
	SetClauses []setClause
	WhereParts []Sqlizer
	OrderBys   []string
	Limit      string
	Offset     string
	Suffixes   exprs
}

type setClause struct {
	column string
	value  interface{}
}

func (d *updateData) Exec() (sql.Result, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	if d.SerializeWith == nil {
		return nil, SerializerNotSet
	}
	return ExecWith(d.RunWith, d, d.SerializeWith)
}

func (d *updateData) Query() (*sql.Rows, error) {
	if d.RunWith == nil {
		return nil, RunnerNotSet
	}
	if d.SerializeWith == nil {
		return nil, SerializerNotSet
	}
	return QueryWith(d.RunWith, d, d.SerializeWith)
}

func (d *updateData) QueryRow() RowScanner {
	if d.RunWith == nil {
		return &Row{err: RunnerNotSet}
	}
	if d.SerializeWith == nil {
		return &Row{err: RunnerNotSet}
	}
	queryRower, ok := d.RunWith.(QueryRower)
	if !ok {
		return &Row{err: RunnerNotQueryRunner}
	}
	return QueryRowWith(queryRower, d, d.SerializeWith)
}

func (d updateData) ToSql() (sqlStr string, args []interface{}, err error) {
	return d.ToSqlWithSerializer(DefaultSerializer{})
}

func (d *updateData) ToSqlWithSerializer(serializer Serializer) (sqlStr string, args []interface{}, err error) {
	if len(d.Table) == 0 {
		err = fmt.Errorf("update statements must specify a table")
		return
	}
	if len(d.SetClauses) == 0 {
		err = fmt.Errorf("update statements must have at least one Set clause")
		return
	}

	return serializer.Update(*d)
}

// Builder

// UpdateBuilder builds SQL UPDATE statements.
type UpdateBuilder builder.Builder

func init() {
	builder.Register(UpdateBuilder{}, updateData{})
}

// Format methods

// PlaceholderFormat sets PlaceholderFormat (e.g. Question or Dollar) for the
// query.
func (b UpdateBuilder) PlaceholderFormat(f PlaceholderFormat) UpdateBuilder {
	return builder.Set(b, "PlaceholderFormat", f).(UpdateBuilder)
}

// Runner methods

// RunWith sets a Runner (like database/sql.DB) to be used with e.g. Exec.
func (b UpdateBuilder) RunWith(runner BaseRunner) UpdateBuilder {
	return setRunWith(b, runner).(UpdateBuilder)
}

// SerializeWith sets a Serializer (that is, db specific writer) to be used with.
func (b UpdateBuilder) SerializeWith(serializer Serializer) UpdateBuilder {
	return setSerializeWith(b, serializer).(UpdateBuilder)
}

// Exec builds and Execs the query with the Runner set by RunWith.
func (b UpdateBuilder) Exec() (sql.Result, error) {
	data := builder.GetStruct(b).(updateData)
	return data.Exec()
}

func (b UpdateBuilder) Query() (*sql.Rows, error) {
	data := builder.GetStruct(b).(updateData)
	return data.Query()
}

func (b UpdateBuilder) QueryRow() RowScanner {
	data := builder.GetStruct(b).(updateData)
	return data.QueryRow()
}

func (b UpdateBuilder) Scan(dest ...interface{}) error {
	return b.QueryRow().Scan(dest...)
}

// SQL methods

// ToSql builds the query into a SQL string and bound args with the default serializer.
func (b UpdateBuilder) ToSql() (sqlStr string, args []interface{}, err error) {
	return b.ToSqlWithSerializer(DefaultSerializer{})
}

// ToSql builds the query into a SQL string and bound args with a specific serializer.
func (b UpdateBuilder) ToSqlWithSerializer(serializer Serializer) (string, []interface{}, error) {
	data := builder.GetStruct(b).(updateData)
	return data.ToSqlWithSerializer(serializer)
}

// Prefix adds an expression to the beginning of the query
func (b UpdateBuilder) Prefix(sql string, args ...interface{}) UpdateBuilder {
	return builder.Append(b, "Prefixes", Expr(sql, args...)).(UpdateBuilder)
}

// Table sets the table to be updated.
func (b UpdateBuilder) Table(table string) UpdateBuilder {
	return builder.Set(b, "Table", table).(UpdateBuilder)
}

// Set adds SET clauses to the query.
func (b UpdateBuilder) Set(column string, value interface{}) UpdateBuilder {
	return builder.Append(b, "SetClauses", setClause{column: column, value: value}).(UpdateBuilder)
}

// SetMap is a convenience method which calls .Set for each key/value pair in clauses.
func (b UpdateBuilder) SetMap(clauses map[string]interface{}) UpdateBuilder {
	keys := make([]string, len(clauses))
	i := 0
	for key := range clauses {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	for _, key := range keys {
		val, _ := clauses[key]
		b = b.Set(key, val)
	}
	return b
}

// Where adds WHERE expressions to the query.
//
// See SelectBuilder.Where for more information.
func (b UpdateBuilder) Where(pred interface{}, args ...interface{}) UpdateBuilder {
	return builder.Append(b, "WhereParts", newWherePart(pred, args...)).(UpdateBuilder)
}

// OrderBy adds ORDER BY expressions to the query.
func (b UpdateBuilder) OrderBy(orderBys ...string) UpdateBuilder {
	return builder.Extend(b, "OrderBys", orderBys).(UpdateBuilder)
}

// Limit sets a LIMIT clause on the query.
func (b UpdateBuilder) Limit(limit uint64) UpdateBuilder {
	return builder.Set(b, "Limit", fmt.Sprintf("%d", limit)).(UpdateBuilder)
}

// Offset sets a OFFSET clause on the query.
func (b UpdateBuilder) Offset(offset uint64) UpdateBuilder {
	return builder.Set(b, "Offset", fmt.Sprintf("%d", offset)).(UpdateBuilder)
}

// Suffix adds an expression to the end of the query
func (b UpdateBuilder) Suffix(sql string, args ...interface{}) UpdateBuilder {
	return builder.Append(b, "Suffixes", Expr(sql, args...)).(UpdateBuilder)
}
