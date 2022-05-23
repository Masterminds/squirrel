package squirrel

import (
	"bytes"
	"errors"

	"github.com/lann/builder"
)

func init() {
	builder.Register(SetBuilder{}, setData{})
}

type setSelect struct {
	op       string
	selector SelectBuilder
}

type setData struct {
	Selects           []*setSelect
	PlaceholderFormat PlaceholderFormat
}

type SetBuilder builder.Builder

func (u SetBuilder) setProp(key string, value interface{}) SetBuilder {
	return builder.Set(u, key, value).(SetBuilder)
}

func (u SetBuilder) ToSql() (sql string, args []interface{}, err error) {
	var (
		selArgs []interface{}
		selSql  string
		sqlBuf  = &bytes.Buffer{}
	)

	data := builder.GetStruct(u).(setData)

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

	sql = sqlBuf.String()

	return sql, args, nil
}

func (u SetBuilder) Union(selector SelectBuilder) SetBuilder {
	selector = selector.PlaceholderFormat(Question)
	return builder.Append(u, "Selects", &setSelect{op: "UNION", selector: selector}).(SetBuilder)
}

func (u SetBuilder) setFirstSelect(selector SelectBuilder) SetBuilder {
	value, _ := builder.Get(selector, "PlaceholderFormat")
	bld := u.setProp("PlaceholderFormat", value)

	selector = selector.PlaceholderFormat(Question)

	return builder.Append(bld, "Selects", &setSelect{op: "", selector: selector}).(SetBuilder)
}

func (u SetBuilder) PlaceholderFormat(fmt PlaceholderFormat) SetBuilder {
	return u.setProp("PlaceholderFormat", fmt)
}

func NewSet(s SelectBuilder) SetBuilder {
	b := SetBuilder{}
	b = b.setFirstSelect(s)
	return b
}

func SelectFromSet(selectBuilder SelectBuilder, set SetBuilder, alias string) SelectBuilder {
	set = set.PlaceholderFormat(Question)
	return builder.Set(selectBuilder, "From", Alias(set, alias)).(SelectBuilder)
}
