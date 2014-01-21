package squirrel

import (
	"fmt"
	"reflect"
	"strings"
)

type Eq map[string]interface{}

type wherePart struct {
	sql  string
	args []interface{}
}

func newWhereParts(pred interface{}, args ...interface{}) []wherePart {
	switch p := pred.(type) {
	case nil:
		return nil
	case Eq:
		return whereEqMap(map[string]interface{}(p))
	case map[string]interface{}:
		return whereEqMap(p)
	case string:
		if len(p) > 0 {
			return []wherePart{{sql: p, args: args}}
		} else {
			return nil
		}
	default:
		panic(fmt.Errorf("expected string-keyed map or string, not %T", pred))
	}
}

func whereEqMap(m map[string]interface{}) (parts []wherePart) {
	for key, val := range m {
		var part wherePart
		if val == nil {
			part.sql = fmt.Sprintf("%s IS NULL", key)
		} else {
			valVal := reflect.ValueOf(val)
			if valVal.Kind() == reflect.Array || valVal.Kind() == reflect.Slice {
				placeholders := make([]string, valVal.Len())
				for i := 0; i < valVal.Len(); i++ {
					placeholders[i] = "?"
					part.args = append(part.args, valVal.Index(i).Interface())
				}
				placeholdersStr := strings.Join(placeholders, ",")
				part.sql = fmt.Sprintf("%s IN (%s)", key, placeholdersStr)
			} else {
				part.sql = fmt.Sprintf("%s = ?", key)
				part.args = []interface{}{val}
			}
		}
		parts = append(parts, part)
	}
	return
}

func wherePartsToSql(parts []wherePart) (string, []interface{}) {
	sqls := make([]string, 0, len(parts))
	var args []interface{}
	for _, part := range parts {
		sqls = append(sqls, part.sql)
		args = append(args, part.args...)
	}
	return strings.Join(sqls, " AND "), args
}
