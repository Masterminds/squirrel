package squirrel

import (
	"fmt"
	"github.com/lann/builder"
)

// Builder

// SubQueryBuilder builds SQL SELECT statements for subquery.
type SubQueryBuilder builder.Builder

func init() {
	builder.Register(SubQueryBuilder{}, selectData{})
}

// SQL methods

// ToSql builds the query into a SQL string and bound args.
func (b SubQueryBuilder) ToSql() (string, []interface{}, error) {
	data := builder.GetStruct(b).(selectData)
	return data.toSubQuerySql()
}

// Prefix adds an expression to the beginning of the query
func (b SubQueryBuilder) Prefix(sql string, args ...interface{}) SubQueryBuilder {
	return builder.Append(b, "Prefixes", Expr(sql, args...)).(SubQueryBuilder)
}

// Distinct adds a DISTINCT clause to the query.
func (b SubQueryBuilder) Distinct() SubQueryBuilder {
	return builder.Set(b, "Distinct", true).(SubQueryBuilder)
}

// Columns adds result columns to the query.
func (b SubQueryBuilder) Columns(columns ...string) SubQueryBuilder {
	var parts []interface{}
	for _, str := range columns {
		parts = append(parts, newPart(str))
	}
	return builder.Extend(b, "Columns", parts).(SubQueryBuilder)
}

// Column adds a result column to the query.
// See SelectBuilder.Column
func (b SubQueryBuilder) Column(column interface{}, args ...interface{}) SubQueryBuilder {
	return builder.Append(b, "Columns", newPart(column, args...)).(SubQueryBuilder)
}

// From sets the FROM clause of the query.
func (b SubQueryBuilder) From(from interface{}) SubQueryBuilder {
	return builder.Set(b, "FromPart", newFromPart(from)).(SubQueryBuilder)
}

// JoinClause adds a join clause to the query.
func (b SubQueryBuilder) JoinClause(join string) SubQueryBuilder {
	return builder.Append(b, "Joins", join).(SubQueryBuilder)
}

// Join adds a JOIN clause to the query.
func (b SubQueryBuilder) Join(join string) SubQueryBuilder {
	return b.JoinClause("JOIN " + join)
}

// LeftJoin adds a LEFT JOIN clause to the query.
func (b SubQueryBuilder) LeftJoin(join string) SubQueryBuilder {
	return b.JoinClause("LEFT JOIN " + join)
}

// RightJoin adds a RIGHT JOIN clause to the query.
func (b SubQueryBuilder) RightJoin(join string) SubQueryBuilder {
	return b.JoinClause("RIGHT JOIN " + join)
}

// Where adds an expression to the WHERE clause of the query.
// See SelectBuilder.Where
func (b SubQueryBuilder) Where(pred interface{}, args ...interface{}) SubQueryBuilder {
	return builder.Append(b, "WhereParts", newWherePart(pred, args...)).(SubQueryBuilder)
}

// GroupBy adds GROUP BY expressions to the query.
func (b SubQueryBuilder) GroupBy(groupBys ...string) SubQueryBuilder {
	return builder.Extend(b, "GroupBys", groupBys).(SubQueryBuilder)
}

// Having adds an expression to the HAVING clause of the query.
//
// See Where.
func (b SubQueryBuilder) Having(pred interface{}, rest ...interface{}) SubQueryBuilder {
	return builder.Append(b, "HavingParts", newWherePart(pred, rest...)).(SubQueryBuilder)
}

// OrderBy adds ORDER BY expressions to the query.
func (b SubQueryBuilder) OrderBy(orderBys ...string) SubQueryBuilder {
	return builder.Extend(b, "OrderBys", orderBys).(SubQueryBuilder)
}

// Limit sets a LIMIT clause on the query.
func (b SubQueryBuilder) Limit(limit uint64) SubQueryBuilder {
	return builder.Set(b, "Limit", fmt.Sprintf("%d", limit)).(SubQueryBuilder)
}

// Offset sets a OFFSET clause on the query.
func (b SubQueryBuilder) Offset(offset uint64) SubQueryBuilder {
	return builder.Set(b, "Offset", fmt.Sprintf("%d", offset)).(SubQueryBuilder)
}

// Suffix adds an expression to the end of the query
func (b SubQueryBuilder) Suffix(sql string, args ...interface{}) SubQueryBuilder {
	return builder.Append(b, "Suffixes", Expr(sql, args...)).(SubQueryBuilder)
}

// Union adds a UNION clause to the query
func (b SubQueryBuilder) Union(query interface{}, args ...interface{}) SubQueryBuilder {
	return builder.Append(b, "Union", newUnionPart(query, args...)).(SubQueryBuilder)
}

// UnionAll adds a UNION ALL clause to the query
func (b SubQueryBuilder) UnionAll(query interface{}, args ...interface{}) SubQueryBuilder {
	return builder.Append(b, "UnionAll", newUnionPart(query, args...)).(SubQueryBuilder)
}

// As adds a AS clause to the query
func (b SubQueryBuilder) As(name string) SubQueryBuilder {
	return builder.Set(b, "As", name).(SubQueryBuilder)
}
