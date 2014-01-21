package squirrel

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"github.com/lann/builder"
)

type selectData struct {
	Distinct    bool
	Columns     []string
	From        string
	WhereParts  []wherePart
	GroupBys    []string
	HavingParts []wherePart
	OrderBys    []string
	Limit       string
	Offset      string
}

func (data *selectData) ToSql() (sqlStr string, args []interface{}, err error) {
	if len(data.Columns) == 0 {
		err = fmt.Errorf("select statements must have at least one result column")
		return
	}

	var sql bytes.Buffer

	sql.WriteString("SELECT ")

	if data.Distinct {
		sql.WriteString("DISTINCT ")
	}

	sql.WriteString(strings.Join(data.Columns, ", "))

	if len(data.From) > 0 {
		sql.WriteString(" FROM ")
		sql.WriteString(data.From)
	}

	if len(data.WhereParts) > 0 {
		sql.WriteString(" WHERE ")
		whereSql, whereArgs := wherePartsToSql(data.WhereParts)
		sql.WriteString(whereSql)
		args = append(args, whereArgs...)
	}

	if len(data.GroupBys) > 0 {
		sql.WriteString(" GROUP BY ")
		sql.WriteString(strings.Join(data.GroupBys, ", "))
	}

	if len(data.HavingParts) > 0 {
		sql.WriteString(" HAVING ")
		havingSql, havingArgs := wherePartsToSql(data.HavingParts)
		sql.WriteString(havingSql)
		args = append(args, havingArgs...)
	}

	if len(data.OrderBys) > 0 {
		sql.WriteString(" ORDER BY ")
		sql.WriteString(strings.Join(data.OrderBys, ", "))
	}

	if len(data.Limit) > 0 {
		sql.WriteString(" LIMIT ")
		sql.WriteString(data.Limit)
	}

	if len(data.Offset) > 0 {
		sql.WriteString(" OFFSET ")
		sql.WriteString(data.Offset)
	}

	sqlStr = sql.String()
	fmt.Println("XXX")
	fmt.Println(sqlStr)
	fmt.Println("XXX")
	return
}

type selectBuilder builder.Builder

var newSelectBuilder = builder.Register(selectBuilder{}, selectData{}).(selectBuilder)

func Select(columns ...string) selectBuilder {
	return newSelectBuilder.Columns(columns...)
}

func (b selectBuilder) ToSql() (sqlStr string, args []interface{}, err error) {
	data := builder.GetStruct(b).(selectData)
	return data.ToSql()
}

func (b selectBuilder) ExecWith(db Execer) (sql.Result, error) {
	return ExecWith(db, b)
}

func (b selectBuilder) QueryWith(db Queryer) (*sql.Rows, error) {
	return QueryWith(db, b)
}

func (b selectBuilder) QueryRowWith(db QueryRower) *Row {
	return QueryRowWith(db, b)
}

// Builder methods

func (b selectBuilder) Distinct() selectBuilder {
	return builder.Set(b, "Distinct", true).(selectBuilder)
}

func (b selectBuilder) Columns(columns ...string) selectBuilder {
	return builder.Extend(b, "Columns", columns).(selectBuilder)
}

func (b selectBuilder) From(from string) selectBuilder {
	return builder.Set(b, "From", from).(selectBuilder)
}

func (b selectBuilder) Where(pred interface{}, rest ...interface{}) selectBuilder {
	return builder.Extend(b, "WhereParts", newWhereParts(pred, rest...)).(selectBuilder)
}

func (b selectBuilder) GroupBy(groupBys ...string) selectBuilder {
	return builder.Extend(b, "GroupBys", groupBys).(selectBuilder)
}

func (b selectBuilder) Having(pred interface{}, rest ...interface{}) selectBuilder {
	return builder.Extend(b, "HavingParts", newWhereParts(pred, rest...)).(selectBuilder)
}

func (b selectBuilder) OrderBy(orderBys ...string) selectBuilder {
	return builder.Extend(b, "OrderBys", orderBys).(selectBuilder)
}

func (b selectBuilder) Limit(limit uint64) selectBuilder {
	return builder.Set(b, "Limit", fmt.Sprintf("%d", limit)).(selectBuilder)
}

func (b selectBuilder) Offset(offset uint64) selectBuilder {
	return builder.Set(b, "Offset", fmt.Sprintf("%d", offset)).(selectBuilder)
}
