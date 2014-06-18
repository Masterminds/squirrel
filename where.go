package squirrel

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

type wherePart struct {
	pred interface{}
	args []interface{}
}

func newWherePart(pred interface{}, args ...interface{}) wherePart {
	return wherePart{pred: pred, args: args}
}

func (p wherePart) ToSql() (sql string, args []interface{}, err error) {
	switch pred := p.pred.(type) {
	case nil:
		// no-op
	case Eq:
		return whereEqMap(map[string]interface{}(pred))
	case map[string]interface{}:
		return whereEqMap(pred)
	case string:
		sql = pred
		args = p.args
	default:
		err = fmt.Errorf("expected string-keyed map or string, not %T", pred)
	}
	return
}

type whereParts []wherePart

func (wps whereParts) AppendToSql(w io.Writer, sep string, args []interface{}) ([]interface{}, error) {
	for i, p := range wps {
		partSql, partArgs, err := p.ToSql()
		if err != nil {
			return nil, err
		} else if len(partSql) == 0 {
			continue
		}

		if i > 0 {
			_, err := io.WriteString(w, sep)
			if err != nil {
				return nil, err
			}
		}

		_, err = io.WriteString(w, partSql)
		if err != nil {
			return nil, err
		}
		args = append(args, partArgs...)
	}
	return args, nil
}

func whereEqMap(m map[string]interface{}) (sql string, args []interface{}, err error) {
	var exprs []string
	for key, val := range m {
		expr := ""
		if val == nil {
			expr = fmt.Sprintf("%s IS NULL", key)
		} else {
			valVal := reflect.ValueOf(val)
			if valVal.Kind() == reflect.Array || valVal.Kind() == reflect.Slice {
				placeholders := make([]string, valVal.Len())
				for i := 0; i < valVal.Len(); i++ {
					placeholders[i] = "?"
					args = append(args, valVal.Index(i).Interface())
				}
				placeholdersStr := strings.Join(placeholders, ",")
				expr = fmt.Sprintf("%s IN (%s)", key, placeholdersStr)
			} else {
				expr = fmt.Sprintf("%s = ?", key)
				args = append(args, val)
			}
		}
		exprs = append(exprs, expr)
	}
	sql = strings.Join(exprs, " AND ")
	return
}
