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
func Expr(sql string, args ...interface{}) expr {
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
	for _, part := range ce {
		switch p := part.(type) {
		case string:
			sql += p
		case Sqlizer:
			pSql, pArgs, err := p.ToSql()
			if err != nil {
				return "", nil, err
			}
			sql += pSql
			args = append(args, pArgs...)
		default:
			return "", nil, fmt.Errorf("%#v is not a string or Sqlizer", part)
		}
	}
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
		sql = fmt.Sprintf("(%s) AS %s", sql, e.alias)
	}
	return
}

type (
	// Eq is syntactic sugar for use with Where/Having/Set methods.
	// Ex:
	//     .Where(Eq{"id": 1})
	Eq         map[string]interface{}
	OptionalEq Eq
)

func (eq Eq) ToSql() (sql string, args []interface{}, err error) {
	return eqToSQL(eq, false, false)
}

func (eq OptionalEq) ToSql() (sql string, args []interface{}, err error) {
	return eqToSQL(eq, false, true)
}

func eqToSQL(eq map[string]interface{}, useNotOpr, optional bool) (sql string, args []interface{}, err error) {
	if len(eq) == 0 {
		// Empty Sql{} evaluates to true.
		sql = sqlTrue
		return
	}

	var (
		exprs       []string
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
	for _, key := range sortedKeys {
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
			expr = fmt.Sprintf("%s %s NULL", key, nullOpr)
		} else {
			if isListType(val) {
				valVal := reflect.ValueOf(val)
				if valVal.Len() == 0 {
					expr = inEmptyExpr
					if args == nil {
						args = []interface{}{}
					}
				} else {
					for i := 0; i < valVal.Len(); i++ {
						args = append(args, valVal.Index(i).Interface())
					}
					expr = fmt.Sprintf("%s %s (%s)", key, inOpr, Placeholders(valVal.Len()))
				}
			} else {
				if optional && isZero(r) {
					continue
				}
				expr = fmt.Sprintf("%s %s ?", key, equalOpr)
				args = append(args, val)
			}
		}
		exprs = append(exprs, expr)
	}
	sql = strings.Join(exprs, " AND ")
	return
}

type (
	// NotEq is syntactic sugar for use with Where/Having/Set methods.
	// Ex:
	//     .Where(NotEq{"id": 1}) == "id <> 1"
	NotEq         Eq
	OptionalNotEq NotEq
)

func (neq NotEq) ToSql() (sql string, args []interface{}, err error) {
	return eqToSQL(neq, true, false)
}

func (neq OptionalNotEq) ToSql() (sql string, args []interface{}, err error) {
	return eqToSQL(neq, true, true)
}

type (
	// Like is syntactic sugar for use with LIKE conditions.
	// Ex:
	//     .Where(Like{"name": "%irrel"})
	Like         map[string]interface{}
	OptionalLike Like
)

func likeToSQL(lk map[string]interface{}, opr string, optional bool) (sql string, args []interface{}, err error) {
	var exprs []string
	for key, val := range lk {
		expr := ""

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
				if optional && (val == "" || val == "%" || val == "%%") {
					continue
				}
				expr = fmt.Sprintf("%s %s ?", key, opr)
				args = append(args, val)
			}
		}
		exprs = append(exprs, expr)
	}
	sql = strings.Join(exprs, " AND ")
	return
}

func (lk Like) ToSql() (sql string, args []interface{}, err error) {
	return likeToSQL(lk, "LIKE", false)
}

func (lk OptionalLike) ToSql() (sql string, args []interface{}, err error) {
	return likeToSQL(lk, "LIKE", true)
}

type (
	// NotLike is syntactic sugar for use with LIKE conditions.
	// Ex:
	//     .Where(NotLike{"name": "%irrel"})
	NotLike         Like
	OptionalNotLike NotLike
)

func (nlk NotLike) ToSql() (sql string, args []interface{}, err error) {
	return likeToSQL(nlk, "NOT LIKE", false)
}

func (nlk OptionalNotLike) ToSql() (sql string, args []interface{}, err error) {
	return likeToSQL(nlk, "NOT LIKE", true)
}

type (
	// ILike is syntactic sugar for use with ILIKE conditions.
	// Ex:
	//    .Where(ILike{"name": "sq%"})
	ILike         Like
	OptionalILike ILike
)

func (ilk ILike) ToSql() (sql string, args []interface{}, err error) {
	return likeToSQL(ilk, "ILIKE", false)
}

func (ilk OptionalILike) ToSql() (sql string, args []interface{}, err error) {
	return likeToSQL(ilk, "ILIKE", true)
}

type (
	// NotILike is syntactic sugar for use with ILIKE conditions.
	// Ex:
	//    .Where(NotILike{"name": "sq%"})
	NotILike         Like
	OptionalNotILike NotILike
)

func (nilk NotILike) ToSql() (sql string, args []interface{}, err error) {
	return likeToSQL(nilk, "NOT ILIKE", false)
}

func (nilk OptionalNotILike) ToSql() (sql string, args []interface{}, err error) {
	return likeToSQL(nilk, "NOT ILIKE", true)
}

type (
	// Lt is syntactic sugar for use with Where/Having/Set methods.
	// Ex:
	//     .Where(Lt{"id": 1})
	Lt         map[string]interface{}
	OptionalLt Lt
)

func ltToSql(lt map[string]interface{}, opposite, orEq, optional bool) (sql string, args []interface{}, err error) {
	var (
		exprs []string
		opr   = "<"
	)

	if opposite {
		opr = ">"
	}

	if orEq {
		opr = fmt.Sprintf("%s%s", opr, "=")
	}

	sortedKeys := getSortedKeys(lt)
	for _, key := range sortedKeys {
		var expr string
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
		if optional {
			if isZero(reflect.ValueOf(val)) {
				continue
			}
		}
		expr = fmt.Sprintf("%s %s ?", key, opr)
		args = append(args, val)

		exprs = append(exprs, expr)
	}
	sql = strings.Join(exprs, " AND ")
	return
}

func (lt Lt) ToSql() (sql string, args []interface{}, err error) {
	return ltToSql(lt, false, false, false)
}

func (lt OptionalLt) ToSql() (sql string, args []interface{}, err error) {
	return ltToSql(lt, false, false, true)
}

type (
	// LtOrEq is syntactic sugar for use with Where/Having/Set methods.
	// Ex:
	//     .Where(LtOrEq{"id": 1}) == "id <= 1"
	LtOrEq         Lt
	OptionalLtOrEq LtOrEq
)

func (ltOrEq LtOrEq) ToSql() (sql string, args []interface{}, err error) {
	return ltToSql(ltOrEq, false, true, false)
}

func (ltOrEq OptionalLtOrEq) ToSql() (sql string, args []interface{}, err error) {
	return ltToSql(ltOrEq, false, true, true)
}

type (
	// Gt is syntactic sugar for use with Where/Having/Set methods.
	// Ex:
	//     .Where(Gt{"id": 1}) == "id > 1"
	Gt         Lt
	OptionalGt Gt
)

func (gt Gt) ToSql() (sql string, args []interface{}, err error) {
	return ltToSql(gt, true, false, false)
}

func (gt OptionalGt) ToSql() (sql string, args []interface{}, err error) {
	return ltToSql(gt, true, false, true)
}

type (
	// GtOrEq is syntactic sugar for use with Where/Having/Set methods.
	// Ex:
	//     .Where(GtOrEq{"id": 1}) == "id >= 1"
	GtOrEq         Lt
	OptionalGtOrEq GtOrEq
)

func (gtOrEq GtOrEq) ToSql() (sql string, args []interface{}, err error) {
	return ltToSql(gtOrEq, true, true, false)
}

func (gtOrEq OptionalGtOrEq) ToSql() (sql string, args []interface{}, err error) {
	return ltToSql(gtOrEq, true, true, true)
}

type conj []Sqlizer

func (c conj) join(sep, defaultExpr string) (sql string, args []interface{}, err error) {
	if len(c) == 0 {
		return defaultExpr, []interface{}{}, nil
	}
	var sqlParts []string
	for _, sqlizer := range c {
		partSQL, partArgs, err := sqlizer.ToSql()
		if err != nil {
			return "", nil, err
		}
		if partSQL != "" {
			sqlParts = append(sqlParts, partSQL)
			args = append(args, partArgs...)
		}
	}
	if len(sqlParts) > 0 {
		sql = fmt.Sprintf("(%s)", strings.Join(sqlParts, sep))
	}
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

// isZero reports whether a value is a zero value
// Including support: Bool, Array, String, Float32, Float64, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr
// Map, Slice, Interface, Struct
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Array, reflect.String:
		return v.Len() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Map, reflect.Slice:
		return v.IsNil() || v.Len() == 0
	case reflect.Interface:
		return v.IsNil()
	case reflect.Invalid:
		return true
	}

	if v.Kind() != reflect.Struct {
		return false
	}

	// Traverse the Struct and only return true
	// if all of its fields return IsZero == true
	n := v.NumField()
	for i := 0; i < n; i++ {
		vf := v.Field(i)
		if !isZero(vf) {
			return false
		}
	}
	return true
}
