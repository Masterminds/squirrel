package squirrel

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

type Eq map[string]any

type wherePart struct {
	sql  string
	args []any
}

func newWherePart(pred any, args ...any) wherePart {
	switch p := pred.(type) {
	case Eq:
		return whereEqMap(map[string]any(p))
	case map[string]any:
		return whereEqMap(p)
	case string:
		return wherePart{sql: p, args: args}
	default:
		log.Panicf("expected string-keyed map or string, not %T", pred)
	}
	return wherePart{}
}

func whereEqMap(m map[string]any) wherePart {
	sqlParts := []string{}
	args := []any{}
	
	for key, val := range m {
		var sqlPart string
		
		valVal := reflect.ValueOf(val)
		if valVal.Kind() == reflect.Array || valVal.Kind() == reflect.Slice {
			placeholders := make([]string, valVal.Len())
			for i := 0; i < valVal.Len(); i++ {
				placeholders[i] = "?"
				args = append(args, valVal.Index(i).Interface())
			}
			sqlPart = fmt.Sprintf("%s IN (%s)", key, strings.Join(placeholders, ","))
		} else {
			args = append(args, val)
			sqlPart = fmt.Sprintf("%s = ?", key)
		}
		
		sqlParts = append(sqlParts, sqlPart)
	}
	
	return wherePart{sql: strings.Join(sqlParts, " AND "), args: args}
}

func wherePartsToSql(parts []wherePart) (string, []any) {
	sqls := make([]string, len(parts))
	var args []any
	for i, part := range parts {
		sqls[i] = part.sql
		args = append(args, part.args...)
	}
	return strings.Join(sqls, " AND "), args
}
