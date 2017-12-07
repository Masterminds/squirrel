package squirrel

import (
	"database/sql/driver"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type expr struct {
	sql  string
	args []interface{}
}

// Expr builds value expressions for InsertBuilder and UpdateBuilder.
//
// Ex:
//     .Values(Expr("FROM_UNIXTIME(?)", t))
func Expr(sql string, args ...interface{}) expr {
	return expr{sql: sql, args: args}
}

func (e expr) ToSql(serializer Serializer) (sql string, args []interface{}, err error) {
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

// aliasExpr helps to alias part of SQL query generated with underlying "expr"
type aliasExpr struct {
	expr  Sqlizer
	alias string
}

// Alias allows to define alias for column in SelectBuilder. Useful when column is
// defined as complex expression like IF or CASE
// Ex:
//		.Column(Alias(caseStmt, "case_column"))
func Alias(expr Sqlizer, alias string) aliasExpr {
	return aliasExpr{expr, alias}
}

func (e aliasExpr) ToSql(serializer Serializer) (sql string, args []interface{}, err error) {
	sql, args, err = e.expr.ToSql(serializer)
	if err == nil {
		sql = fmt.Sprintf("(%s) AS %s", sql, e.alias)
	}
	return
}

// Eq is syntactic sugar for use with Where/Having/Set methods.
// Ex:
//     .Where(Eq{"id": 1})
type Eq map[string]interface{}

func (eq Eq) toSql(useNotOpr bool, serializer Serializer) (sql string, args []interface{}, err error) {
	return serializer.EQ(eq, useNotOpr)
}

func (eq Eq) ToSql(serializer Serializer) (sql string, args []interface{}, err error) {
	return eq.toSql(false, serializer)
}

// NotEq is syntactic sugar for use with Where/Having/Set methods.
// Ex:
//     .Where(NotEq{"id": 1}) == "id <> 1"
type NotEq Eq

func (neq NotEq) ToSql(serializer Serializer) (sql string, args []interface{}, err error) {
	return Eq(neq).toSql(true, serializer)
}

// Lt is syntactic sugar for use with Where/Having/Set methods.
// Ex:
//     .Where(Lt{"id": 1})
type Lt map[string]interface{}

func (lt Lt) toSql(opposite, orEq bool, serializer Serializer) (sql string, args []interface{}, err error) {
	return serializer.LT(lt, opposite, orEq)
}

func (lt Lt) ToSql(serializer Serializer) (sql string, args []interface{}, err error) {
	return lt.toSql(false, false, serializer)
}

// LtOrEq is syntactic sugar for use with Where/Having/Set methods.
// Ex:
//     .Where(LtOrEq{"id": 1}) == "id <= 1"
type LtOrEq Lt

func (ltOrEq LtOrEq) ToSql(serializer Serializer) (sql string, args []interface{}, err error) {
	return Lt(ltOrEq).toSql(false, true, serializer)
}

// Gt is syntactic sugar for use with Where/Having/Set methods.
// Ex:
//     .Where(Gt{"id": 1}) == "id > 1"
type Gt Lt

func (gt Gt) ToSql(serializer Serializer) (sql string, args []interface{}, err error) {
	return Lt(gt).toSql(true, false, serializer)
}

// GtOrEq is syntactic sugar for use with Where/Having/Set methods.
// Ex:
//     .Where(GtOrEq{"id": 1}) == "id >= 1"
type GtOrEq Lt

func (gtOrEq GtOrEq) ToSql(serializer Serializer) (sql string, args []interface{}, err error) {
	return Lt(gtOrEq).toSql(true, true, serializer)
}

type conj []Sqlizer

func (c conj) join(sep string, serializer Serializer) (sql string, args []interface{}, err error) {
	var sqlParts []string
	for _, sqlizer := range c {
		partSql, partArgs, err := sqlizer.ToSql(serializer)
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

func (a And) ToSql(serializer Serializer) (string, []interface{}, error) {
	return conj(a).join(" AND ", serializer)
}

type Or conj

func (o Or) ToSql(serializer Serializer) (string, []interface{}, error) {
	return conj(o).join(" OR ", serializer)
}

func isListType(val interface{}) bool {
	if driver.IsValue(val) {
		return false
	}
	valVal := reflect.ValueOf(val)
	return valVal.Kind() == reflect.Array || valVal.Kind() == reflect.Slice
}
