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

func (d *setData) ToSql() (sql string, args []interface{}, err error) {
	var (
		selArgs []interface{}
		selSql  string
		sqlBuf  = &bytes.Buffer{}
	)

	if len(d.Selects) == 0 {
		err = errors.New("require a minimum of 1 select clause in SetBuilder")
		return sql, args, err
	}

	for index, selector := range d.Selects {
		selSql, selArgs, err = selector.selector.ToSql()

		if err != nil {
			return sql, args, err
		}

		if index == 0 {
			sqlBuf.WriteString(selSql)
		} else {
			sqlBuf.WriteString(" " + selector.op + " " + selSql)
		}

		args = append(args, selArgs...)
	}

	return sqlBuf.String(), args, err
}

func (b SetBuilder) setProp(key string, value interface{}) SetBuilder {
	return builder.Set(b, key, value).(SetBuilder)
}

func (b SetBuilder) ToSql() (sql string, args []interface{}, err error) {
	data := builder.GetStruct(b).(setData)
	return data.ToSql()
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
