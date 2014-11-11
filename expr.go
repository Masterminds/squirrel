package squirrel

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

type expr struct {
	sql string
	args []interface{}
}

// Expr builds value expressions for InsertBuilder and UpdateBuilder.
//
// Ex:
//     .Values(Expr("FROM_UNIXTIME(?)", t))
func Expr(sql string, args ...interface{}) expr {
	return expr{sql: sql, args: args}
}

func (e expr) ToSql() (sql string, args []interface{}, err error) {
	return e.sql, e.args, nil
}

type exprs []expr

func (es exprs) AppendToSql(w io.Writer, sep string, args []interface{}) ([]interface{}, error) {
	for i, e := range es {
		if i > 0 {
			_, err := io.WriteString(w, sep)
			if err != nil {
				return nil, err
			}
		}
		_, err := io.WriteString(w, e.sql)
		if err != nil {
			return nil, err
		}
		args = append(args, e.args...)
	}
	return args, nil
}

// Eq is syntactic sugar for use with Where/Having/Set methods.
// Ex:
//     .Where(Eq{"id": 1})
type Eq map[string]interface{}

func (eq Eq) ToSql() (sql string, args []interface{}, err error) {
	var exprs []string
	for key, val := range eq {
		expr := ""
		if val == nil {
			expr = fmt.Sprintf("%s IS NULL", key)
		} else {
			valVal := reflect.ValueOf(val)
			if valVal.Kind() == reflect.Array || valVal.Kind() == reflect.Slice {
				if valVal.Len() == 0 {
					err = fmt.Errorf("equality condition must contain at least one paramater")
					return
				}
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

type conj []Sqlizer

func (c conj) join(sep string) (sql string, args []interface{}, err error) {
	var sqlParts []string
	for _, sqlizer := range c {
		partSql, partArgs, err := sqlizer.ToSql()
		if err != nil {
			return "", nil, err
		}
		if partSql != "" {
			sqlParts = append(sqlParts, partSql)
			args = append(args, partArgs...)
		}
	}
	if len(sqlParts) > 0 {
		sql = fmt.Sprintf("(%s)", strings.Join(sqlParts, sep))
	}
	return
}

type And conj

func (a And) ToSql() (string, []interface{}, error) {
	return conj(a).join(" AND ")
}

type Or conj

func (o Or) ToSql() (string, []interface{}, error) {
	return conj(o).join(" OR ")
}
