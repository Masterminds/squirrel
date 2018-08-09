package squirrel

import (
	"fmt"

	"github.com/lann/builder"
)

type unionPart part

func newUnionPart(pred interface{}, args ...interface{}) Sqlizer {
	return &unionPart{pred: pred, args: args}
}
func (p unionPart) ToSql() (sqlStr string, args []interface{}, err error) {
	switch pred := p.pred.(type) {
	case SelectBuilder:
		entity := builder.GetStruct(pred).(selectData)
		sqlStr, args, err = entity.ToSql()
		fmt.Println(sqlStr)
	case string:
		sqlStr = pred
		args = p.args
	default:
		err = fmt.Errorf("expected string or SelectBuilder, not %T", pred)
	}
	return
}
