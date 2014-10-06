package squirrel

import (
	"fmt"
)

type fromPart part

func newFromPart(pred interface{}) Sqlizer {
	return &fromPart{pred: pred, args: []interface{}{}}
}

func (p fromPart) ToSql() (sql string, args []interface{}, err error) {
	switch pred := p.pred.(type) {
	case SubQueryBuilder:
		sql, args, err = pred.ToSql()
	case string:
		sql = pred
		args = []interface{}{}
	default:
		err = fmt.Errorf("expected string or SelectBuilder, not %T", pred)
	}
	return
}
