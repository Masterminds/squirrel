package squirrel

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"github.com/lann/builder"
)

type selectData struct {
	RunWith     Runner
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

var RunnerNotSet = fmt.Errorf("cannot run; no Runner set (RunWith)")

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

func (d *selectData) QueryRow() *Row  {
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

	var sql bytes.Buffer

	sql.WriteString("SELECT ")

	if d.Distinct {
		sql.WriteString("DISTINCT ")
	}

	sql.WriteString(strings.Join(d.Columns, ", "))

	if len(d.From) > 0 {
		sql.WriteString(" FROM ")
		sql.WriteString(d.From)
	}

	if len(d.WhereParts) > 0 {
		sql.WriteString(" WHERE ")
		whereSql, whereArgs := wherePartsToSql(d.WhereParts)
		sql.WriteString(whereSql)
		args = append(args, whereArgs...)
	}

	if len(d.GroupBys) > 0 {
		sql.WriteString(" GROUP BY ")
		sql.WriteString(strings.Join(d.GroupBys, ", "))
	}

	if len(d.HavingParts) > 0 {
		sql.WriteString(" HAVING ")
		havingSql, havingArgs := wherePartsToSql(d.HavingParts)
		sql.WriteString(havingSql)
		args = append(args, havingArgs...)
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

	sqlStr = sql.String()
	return
}

// Builder

type selectBuilder builder.Builder

var newSelectBuilder = builder.Register(selectBuilder{}, selectData{}).(selectBuilder)

func Select(columns ...string) selectBuilder {
	return newSelectBuilder.Columns(columns...)
}

func selectWith(runner Runner, columns ...string) selectBuilder {
	return Select(columns...).RunWith(runner)
}

// Runner methods

func (b selectBuilder) RunWith(runner Runner) selectBuilder {
	return builder.Set(b, "RunWith", runner).(selectBuilder)
}

func (b selectBuilder) Exec() (sql.Result, error) {
	data := builder.GetStruct(b).(selectData)
	return data.Exec()
}

func (b selectBuilder) Query() (*sql.Rows, error) {
	data := builder.GetStruct(b).(selectData)
	return data.Query()
}

func (b selectBuilder) QueryRow() *Row  {
	data := builder.GetStruct(b).(selectData)
	return data.QueryRow()
}

func (b selectBuilder) Scan(dest ...interface{}) error {
	return b.QueryRow().Scan(dest...)
}

// SQL methods

func (b selectBuilder) ToSql() (string, []interface{}, error) {
	data := builder.GetStruct(b).(selectData)
	return data.ToSql()
}

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
