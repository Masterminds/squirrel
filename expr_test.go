package squirrel

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcatExpr(t *testing.T) {
	b := ConcatExpr("COALESCE(name,", Expr("CONCAT(?,' ',?)", "f", "l"), ")")
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "COALESCE(name,CONCAT(?,' ',?))"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{"f", "l"}
	assert.Equal(t, expectedArgs, args)
}

func TestConcatExprBadType(t *testing.T) {
	b := ConcatExpr("prefix", 123, "suffix")
	_, _, err := b.ToSql()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "123 is not")
}

func TestEqToSql(t *testing.T) {
	b := Eq{"id": 1}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "id = ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1}
	assert.Equal(t, expectedArgs, args)
}

func TestEqEmptyToSql(t *testing.T) {
	sql, args, err := Eq{}.ToSql()
	assert.NoError(t, err)

	expectedSql := "(1=1)"
	assert.Equal(t, expectedSql, sql)
	assert.Empty(t, args)
}

func TestEqInToSql(t *testing.T) {
	b := Eq{"id": []int{1, 2, 3}}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "id IN (?,?,?)"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1, 2, 3}
	assert.Equal(t, expectedArgs, args)
}

func TestNotEqToSql(t *testing.T) {
	b := NotEq{"id": 1}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "id <> ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1}
	assert.Equal(t, expectedArgs, args)
}

func TestEqNotInToSql(t *testing.T) {
	b := NotEq{"id": []int{1, 2, 3}}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "id NOT IN (?,?,?)"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1, 2, 3}
	assert.Equal(t, expectedArgs, args)
}

func TestEqInEmptyToSql(t *testing.T) {
	b := Eq{"id": []int{}}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "(1=0)"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{}
	assert.Equal(t, expectedArgs, args)
}

func TestNotEqInEmptyToSql(t *testing.T) {
	b := NotEq{"id": []int{}}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "(1=1)"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{}
	assert.Equal(t, expectedArgs, args)
}

func TestEqBytesToSql(t *testing.T) {
	b := Eq{"id": []byte("test")}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "id = ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{[]byte("test")}
	assert.Equal(t, expectedArgs, args)
}

func TestLtToSql(t *testing.T) {
	b := Lt{"id": 1}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "id < ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1}
	assert.Equal(t, expectedArgs, args)
}

func TestLtOrEqToSql(t *testing.T) {
	b := LtOrEq{"id": 1}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "id <= ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1}
	assert.Equal(t, expectedArgs, args)
}

func TestGtToSql(t *testing.T) {
	b := Gt{"id": 1}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "id > ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1}
	assert.Equal(t, expectedArgs, args)
}

func TestGtOrEqToSql(t *testing.T) {
	b := GtOrEq{"id": 1}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "id >= ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1}
	assert.Equal(t, expectedArgs, args)
}

func TestExprNilToSql(t *testing.T) {
	var b Sqlizer
	b = NotEq{"name": nil}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)
	assert.Empty(t, args)

	expectedSql := "name IS NOT NULL"
	assert.Equal(t, expectedSql, sql)

	b = Eq{"name": nil}
	sql, args, err = b.ToSql()
	assert.NoError(t, err)
	assert.Empty(t, args)

	expectedSql = "name IS NULL"
	assert.Equal(t, expectedSql, sql)
}

func TestNullTypeString(t *testing.T) {
	var b Sqlizer
	var name sql.NullString

	b = Eq{"name": name}
	sql, args, err := b.ToSql()

	assert.NoError(t, err)
	assert.Empty(t, args)
	assert.Equal(t, "name IS NULL", sql)

	name.Scan("Name")
	b = Eq{"name": name}
	sql, args, err = b.ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"Name"}, args)
	assert.Equal(t, "name = ?", sql)
}

func TestNullTypeInt64(t *testing.T) {
	var userID sql.NullInt64
	userID.Scan(nil)
	b := Eq{"user_id": userID}
	sql, args, err := b.ToSql()

	assert.NoError(t, err)
	assert.Empty(t, args)
	assert.Equal(t, "user_id IS NULL", sql)

	userID.Scan(int64(10))
	b = Eq{"user_id": userID}
	sql, args, err = b.ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{int64(10)}, args)
	assert.Equal(t, "user_id = ?", sql)
}

func TestNilPointer(t *testing.T) {
	var name *string = nil
	eq := Eq{"name": name}
	sql, args, err := eq.ToSql()

	assert.NoError(t, err)
	assert.Empty(t, args)
	assert.Equal(t, "name IS NULL", sql)

	neq := NotEq{"name": name}
	sql, args, err = neq.ToSql()

	assert.NoError(t, err)
	assert.Empty(t, args)
	assert.Equal(t, "name IS NOT NULL", sql)

	var ids *[]int = nil
	eq = Eq{"id": ids}
	sql, args, err = eq.ToSql()
	assert.NoError(t, err)
	assert.Empty(t, args)
	assert.Equal(t, "id IS NULL", sql)

	neq = NotEq{"id": ids}
	sql, args, err = neq.ToSql()
	assert.NoError(t, err)
	assert.Empty(t, args)
	assert.Equal(t, "id IS NOT NULL", sql)

	var ida *[3]int = nil
	eq = Eq{"id": ida}
	sql, args, err = eq.ToSql()
	assert.NoError(t, err)
	assert.Empty(t, args)
	assert.Equal(t, "id IS NULL", sql)

	neq = NotEq{"id": ida}
	sql, args, err = neq.ToSql()
	assert.NoError(t, err)
	assert.Empty(t, args)
	assert.Equal(t, "id IS NOT NULL", sql)

}

func TestNotNilPointer(t *testing.T) {
	c := "Name"
	name := &c
	eq := Eq{"name": name}
	sql, args, err := eq.ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"Name"}, args)
	assert.Equal(t, "name = ?", sql)

	neq := NotEq{"name": name}
	sql, args, err = neq.ToSql()

	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"Name"}, args)
	assert.Equal(t, "name <> ?", sql)

	s := []int{1, 2, 3}
	ids := &s
	eq = Eq{"id": ids}
	sql, args, err = eq.ToSql()
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{1, 2, 3}, args)
	assert.Equal(t, "id IN (?,?,?)", sql)

	neq = NotEq{"id": ids}
	sql, args, err = neq.ToSql()
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{1, 2, 3}, args)
	assert.Equal(t, "id NOT IN (?,?,?)", sql)

	a := [3]int{1, 2, 3}
	ida := &a
	eq = Eq{"id": ida}
	sql, args, err = eq.ToSql()
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{1, 2, 3}, args)
	assert.Equal(t, "id IN (?,?,?)", sql)

	neq = NotEq{"id": ida}
	sql, args, err = neq.ToSql()
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{1, 2, 3}, args)
	assert.Equal(t, "id NOT IN (?,?,?)", sql)
}

func TestEmptyAndToSql(t *testing.T) {
	sql, args, err := And{}.ToSql()
	assert.NoError(t, err)

	expectedSql := "(1=1)"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{}
	assert.Equal(t, expectedArgs, args)
}

func TestEmptyOrToSql(t *testing.T) {
	sql, args, err := Or{}.ToSql()
	assert.NoError(t, err)

	expectedSql := "(1=0)"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{}
	assert.Equal(t, expectedArgs, args)
}

func TestLikeToSql(t *testing.T) {
	b := Like{"name": "%irrel"}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "name LIKE ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{"%irrel"}
	assert.Equal(t, expectedArgs, args)
}

func TestNotLikeToSql(t *testing.T) {
	b := NotLike{"name": "%irrel"}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "name NOT LIKE ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{"%irrel"}
	assert.Equal(t, expectedArgs, args)
}

func TestILikeToSql(t *testing.T) {
	b := ILike{"name": "sq%"}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "name ILIKE ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{"sq%"}
	assert.Equal(t, expectedArgs, args)
}

func TestNotILikeToSql(t *testing.T) {
	b := NotILike{"name": "sq%"}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "name NOT ILIKE ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{"sq%"}
	assert.Equal(t, expectedArgs, args)
}

func TestSqlEqOrder(t *testing.T) {
	b := Eq{"a": 1, "b": 2, "c": 3}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "a = ? AND b = ? AND c = ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1, 2, 3}
	assert.Equal(t, expectedArgs, args)
}

func TestSqlLtOrder(t *testing.T) {
	b := Lt{"a": 1, "b": 2, "c": 3}
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "a < ? AND b < ? AND c < ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1, 2, 3}
	assert.Equal(t, expectedArgs, args)
}

func TestExprEscaped(t *testing.T) {
	b := Expr("count(??)", Expr("x"))
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "count(??)"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{Expr("x")}
	assert.Equal(t, expectedArgs, args)
}

func TestExprRecursion(t *testing.T) {
	{
		b := Expr("count(?)", Expr("nullif(a,?)", "b"))
		sql, args, err := b.ToSql()
		assert.NoError(t, err)

		expectedSql := "count(nullif(a,?))"
		assert.Equal(t, expectedSql, sql)

		expectedArgs := []interface{}{"b"}
		assert.Equal(t, expectedArgs, args)
	}
	{
		b := Expr("extract(? from ?)", Expr("epoch"), "2001-02-03")
		sql, args, err := b.ToSql()
		assert.NoError(t, err)

		expectedSql := "extract(epoch from ?)"
		assert.Equal(t, expectedSql, sql)

		expectedArgs := []interface{}{"2001-02-03"}
		assert.Equal(t, expectedArgs, args)
	}
	{
		b := Expr("JOIN t1 ON ?", And{Eq{"id": 1}, Expr("NOT c1"), Expr("? @@ ?", "x", "y")})
		sql, args, err := b.ToSql()
		assert.NoError(t, err)

		expectedSql := "JOIN t1 ON (id = ? AND NOT c1 AND ? @@ ?)"
		assert.Equal(t, expectedSql, sql)

		expectedArgs := []interface{}{1, "x", "y"}
		assert.Equal(t, expectedArgs, args)
	}
}
