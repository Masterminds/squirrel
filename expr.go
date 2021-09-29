package squirrel

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

const (
	// Portable true/false literals.
	sqlTrue  = "(1=1)"
	sqlFalse = "(1=0)"
)

type expr struct {
	sql  string
	args []interface{}
}

// Expr builds an expression from a SQL fragment and arguments.
//
// Ex:
//     Expr("FROM_UNIXTIME(?)", t)
func Expr(sql string, args ...interface{}) Sqlizer {
	return expr{sql: sql, args: args}
}

func (e expr) ToSql() (sql string, args []interface{}, err error) {
	simple := true
	for _, arg := range e.args {
		if _, ok := arg.(Sqlizer); ok {
			simple = false
		}
	}
	if simple {
		return e.sql, e.args, nil
	}

	buf := &bytes.Buffer{}
	ap := e.args
	sp := e.sql

	var isql string
	var iargs []interface{}

	for err == nil && len(ap) > 0 && len(sp) > 0 {
		i := strings.Index(sp, "?")
		if i < 0 {
			// no more placeholders
			break
		}
		if len(sp) > i+1 && sp[i+1:i+2] == "?" {
			// escaped "??"; append it and step past
			buf.WriteString(sp[:i+2])
			sp = sp[i+2:]
			continue
		}

		if as, ok := ap[0].(Sqlizer); ok {
			// sqlizer argument; expand it and append the result
			isql, iargs, err = as.ToSql()
			buf.WriteString(sp[:i])
			buf.WriteString(isql)
			args = append(args, iargs...)
		} else {
			// normal argument; append it and the placeholder
			buf.WriteString(sp[:i+1])
			args = append(args, ap[0])
		}

		// step past the argument and placeholder
		ap = ap[1:]
		sp = sp[i+1:]
	}

	// append the remaining sql and arguments
	buf.WriteString(sp)
	return buf.String(), append(args, ap...), err
}

type concatExpr []interface{}

func (ce concatExpr) ToSql() (sql string, args []interface{}, err error) {
	var b strings.Builder
	b.WriteString(sql)
	for _, part := range ce {
		switch p := part.(type) {
		case string:
			b.WriteString(p)
		case Sqlizer:
			pSql, pArgs, err := p.ToSql()
			if err != nil {
				return "", nil, err
			}
			b.WriteString(pSql)
			args = append(args, pArgs...)
		default:
			return "", nil, fmt.Errorf("%#v is not a string or Sqlizer", part)
		}
	}
	sql = b.String()
	return
}

// ConcatExpr builds an expression by concatenating strings and other expressions.
//
// Ex:
//     name_expr := Expr("CONCAT(?, ' ', ?)", firstName, lastName)
//     ConcatExpr("COALESCE(full_name,", name_expr, ")")
func ConcatExpr(parts ...interface{}) concatExpr {
	return concatExpr(parts)
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

func (e aliasExpr) ToSql() (sql string, args []interface{}, err error) {
	sql, args, err = e.expr.ToSql()
	if err == nil {
		sql = "(" + sql + ") AS " + e.alias
	}
	return
}

// Eq is syntactic sugar for use with Where/Having/Set methods.
type Eq map[string]interface{}

func (eq Eq) toSQL(useNotOpr bool) (sql string, args []interface{}, err error) {
	if len(eq) == 0 {
		// Empty Sql{} evaluates to true.
		sql = sqlTrue
		return
	}

	var (
		equalOpr    = "="
		inOpr       = "IN"
		nullOpr     = "IS"
		inEmptyExpr = sqlFalse
	)

	if useNotOpr {
		equalOpr = "<>"
		inOpr = "NOT IN"
		nullOpr = "IS NOT"
		inEmptyExpr = sqlTrue
	}

	sortedKeys := getSortedKeys(eq)
	var b strings.Builder
	for i, key := range sortedKeys {
		var expr string
		val := eq[key]

		switch v := val.(type) {
		case driver.Valuer:
			if val, err = v.Value(); err != nil {
				return
			}
		}

		r := reflect.ValueOf(val)
		if r.Kind() == reflect.Ptr {
			if r.IsNil() {
				val = nil
			} else {
				val = r.Elem().Interface()
			}
		}

		if val == nil {
			expr = key + " " + nullOpr + " NULL"
		} else {
			if isListType(val) {
				valVal := reflect.ValueOf(val)
				valLen := valVal.Len()
				if valLen == 0 {
					expr = inEmptyExpr
					if args == nil {
						args = []interface{}{}
					}
				} else {
					for i := 0; i < valLen; i++ {
						args = append(args, valVal.Index(i).Interface())
					}
					expr = key + " " + inOpr + " (" + Placeholders(valLen) + ")"
				}
			} else {
				expr = key + " " + equalOpr + " ?"
				args = append(args, val)
			}
		}
		b.WriteString(expr)
		if i != len(sortedKeys)-1 {
			b.WriteString(" AND ")
		}
	}
	sql = b.String()
	return
}

func (eq Eq) ToSql() (sql string, args []interface{}, err error) {
	return eq.toSQL(false)
}

// NotEq is syntactic sugar for use with Where/Having/Set methods.
// Ex:
//     .Where(NotEq{"id": 1}) == "id <> 1"
type NotEq Eq

func (neq NotEq) ToSql() (sql string, args []interface{}, err error) {
	return Eq(neq).toSQL(true)
}

// Like is syntactic sugar for use with LIKE conditions.
// Ex:
//     .Where(Like{"name": "%irrel"})
type Like map[string]interface{}

func (lk Like) toSql(opr string) (sql string, args []interface{}, err error) {
	var b strings.Builder
	i := 0
	for key, val := range lk {
		switch v := val.(type) {
		case driver.Valuer:
			if val, err = v.Value(); err != nil {
				return
			}
		}

		if val == nil {
			err = fmt.Errorf("cannot use null with like operators")
			return
		} else {
			if isListType(val) {
				err = fmt.Errorf("cannot use array or slice with like operators")
				return
			} else {
				b.WriteString(key + " " + opr + " ?")
				args = append(args, val)
			}
		}
		if i != len(lk)-1 {
			b.WriteString(" AND ")
		}
		i++
	}
	sql = b.String()
	return
}

func (lk Like) ToSql() (sql string, args []interface{}, err error) {
	return lk.toSql("LIKE")
}

// NotLike is syntactic sugar for use with LIKE conditions.
// Ex:
//     .Where(NotLike{"name": "%irrel"})
type NotLike Like

func (nlk NotLike) ToSql() (sql string, args []interface{}, err error) {
	return Like(nlk).toSql("NOT LIKE")
}

// ILike is syntactic sugar for use with ILIKE conditions.
// Ex:
//    .Where(ILike{"name": "sq%"})
type ILike Like

func (ilk ILike) ToSql() (sql string, args []interface{}, err error) {
	return Like(ilk).toSql("ILIKE")
}

// NotILike is syntactic sugar for use with ILIKE conditions.
// Ex:
//    .Where(NotILike{"name": "sq%"})
type NotILike Like

func (nilk NotILike) ToSql() (sql string, args []interface{}, err error) {
	return Like(nilk).toSql("NOT ILIKE")
}

// Lt is syntactic sugar for use with Where/Having/Set methods.
// Ex:
//     .Where(Lt{"id": 1})
type Lt map[string]interface{}

func (lt Lt) toSql(opposite, orEq bool) (sql string, args []interface{}, err error) {
	var (
		opr = "<"
	)

	if opposite {
		opr = ">"
	}

	if orEq {
		opr = opr + "="
	}

	sortedKeys := getSortedKeys(lt)
	var b strings.Builder
	for i, key := range sortedKeys {
		val := lt[key]

		switch v := val.(type) {
		case driver.Valuer:
			if val, err = v.Value(); err != nil {
				return
			}
		}

		if val == nil {
			err = fmt.Errorf("cannot use null with less than or greater than operators")
			return
		}
		if isListType(val) {
			err = fmt.Errorf("cannot use array or slice with less than or greater than operators")
			return
		}
		b.WriteString(key + " " + opr + " ?")
		args = append(args, val)

		if i != len(sortedKeys)-1 {
			b.WriteString(" AND ")
		}
	}
	sql = b.String()
	return
}

func (lt Lt) ToSql() (sql string, args []interface{}, err error) {
	return lt.toSql(false, false)
}

// LtOrEq is syntactic sugar for use with Where/Having/Set methods.
// Ex:
//     .Where(LtOrEq{"id": 1}) == "id <= 1"
type LtOrEq Lt

func (ltOrEq LtOrEq) ToSql() (sql string, args []interface{}, err error) {
	return Lt(ltOrEq).toSql(false, true)
}

// Gt is syntactic sugar for use with Where/Having/Set methods.
// Ex:
//     .Where(Gt{"id": 1}) == "id > 1"
type Gt Lt

func (gt Gt) ToSql() (sql string, args []interface{}, err error) {
	return Lt(gt).toSql(true, false)
}

// GtOrEq is syntactic sugar for use with Where/Having/Set methods.
// Ex:
//     .Where(GtOrEq{"id": 1}) == "id >= 1"
type GtOrEq Lt

func (gtOrEq GtOrEq) ToSql() (sql string, args []interface{}, err error) {
	return Lt(gtOrEq).toSql(true, true)
}

type conj []Sqlizer

func (c conj) join(sep, defaultExpr string) (sql string, args []interface{}, err error) {
	if len(c) == 0 {
		return defaultExpr, []interface{}{}, nil
	}
	var b strings.Builder
	hasParts := false
	for i, sqlizer := range c {
		partSQL, partArgs, err := sqlizer.ToSql()
		if err != nil {
			return "", nil, err
		}
		if partSQL != "" {
			if !hasParts {
				b.WriteString("(")
				hasParts = true
			}
			b.WriteString(partSQL)
			args = append(args, partArgs...)
			if i != len(c)-1 {
				b.WriteString(sep)
			}
		}
	}

	if hasParts {
		b.WriteString(")")
	}
	sql = b.String()
	return
}

// And conjunction Sqlizers
type And conj

func (a And) ToSql() (string, []interface{}, error) {
	return conj(a).join(" AND ", sqlTrue)
}

// Or conjunction Sqlizers
type Or conj

func (o Or) ToSql() (string, []interface{}, error) {
	return conj(o).join(" OR ", sqlFalse)
}

func getSortedKeys(exp map[string]interface{}) []string {
	sortedKeys := make([]string, 0, len(exp))
	for k := range exp {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

func isListType(val interface{}) bool {
	if driver.IsValue(val) {
		return false
	}
	valVal := reflect.ValueOf(val)
	return valVal.Kind() == reflect.Array || valVal.Kind() == reflect.Slice
}
