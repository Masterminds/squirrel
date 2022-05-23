package squirrel

import (
	"bytes"
	"errors"

	"github.com/lann/builder"
)

type SetBuilder builder.Builder

func init() {
	builder.Register(SetBuilder{}, setData{})
}

type setData struct {
	Selects           []*setSelect
	PlaceholderFormat PlaceholderFormat
}

type setSelect struct {
	op       string
	selector SelectBuilder
}

func (b SetBuilder) setProp(key string, value interface{}) SetBuilder {
	return builder.Set(b, key, value).(SetBuilder)
}

func (b SetBuilder) ToSql() (sql string, args []interface{}, err error) {
	var (
		selArgs []interface{}
		selSql  string
		sqlBuf  = &bytes.Buffer{}
	)

	data := builder.GetStruct(b).(setData)

	if len(data.Selects) == 0 {
		err = errors.New("require a minimum of 1 select clause in SetBuilder")
		return sql, args, err
	}

	for index, selector := range data.Selects {
		selSql, selArgs, err = selector.selector.ToSql()

		if err != nil {
			return sql, args, err
		}

		if index == 0 {
			sqlBuf.WriteString(selSql) // no operator for first select-clause
		} else {
			sqlBuf.WriteString(" " + selector.op + " ( " + selSql + " ) ")
		}

		args = append(args, selArgs...)
	}

	sql, err = data.PlaceholderFormat.ReplacePlaceholders(sqlBuf.String())

	return sql, args, err
}

func (b SetBuilder) PlaceholderFormat(fmt PlaceholderFormat) SetBuilder {
	return b.setProp("PlaceholderFormat", fmt)
}

func (b SetBuilder) Union(selector SelectBuilder) SetBuilder {
	selector = selector.PlaceholderFormat(Question)
	return builder.Append(b, "Selects", &setSelect{op: "UNION", selector: selector}).(SetBuilder)
}

func (b SetBuilder) setFirstSelect(selector SelectBuilder) SetBuilder {
	selector = selector.PlaceholderFormat(Question)
	return builder.Append(b, "Selects", &setSelect{op: "", selector: selector}).(SetBuilder)
}

func SelectFromSet(selectBuilder SelectBuilder, set SetBuilder, alias string) SelectBuilder {
	set = set.PlaceholderFormat(Question)
	return builder.Set(selectBuilder, "From", Alias(set, alias)).(SelectBuilder)
}
