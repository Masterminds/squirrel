package squirrel

import (
	"fmt"
)

type wherePart part

func newWherePart(pred interface{}, args ...interface{}) Sqlizer {
	return &wherePart{pred: pred, args: args}
}

func (p wherePart) ToSql() (sql string, args []interface{}, err error) {
	switch pred := p.pred.(type) {
	case nil:
		// no-op
	case rawSqlizer:
		return pred.toSqlRaw()
	case Sqlizer:
		return pred.ToSql()
	case map[string]interface{}:
		return Eq(pred).ToSql()
	case string:
		sql = pred
		args = p.args
	default:
		err = fmt.Errorf("expected string-keyed map or string, not %T", pred)
	}
	return
}

func ensureWherePartIsNonEmptyMap(whereParts []Sqlizer) error {
	if len(whereParts) == 1 {
		if wp, ok := whereParts[0].(*wherePart); ok {
			if m, ok := wp.pred.(map[string]interface{}); ok {
				if len(m) == 0 {
					return ErrUpdateOrDeleteWithNilMap
				}
			}
		}
	}
	return nil
}
