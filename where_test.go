package squirrel

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWherePartsAppendToSql(t *testing.T) {
	parts := []Sqlizer{
		newWherePart("x = ?", 1),
		newWherePart(nil),
		newWherePart(Eq{"y": 2}),
	}
	sql := &bytes.Buffer{}
	args, _ := appendToSql(parts, sql, " AND ", []interface{}{})
	assert.Equal(t, "x = ? AND y = ?", sql.String())
	assert.Equal(t, []interface{}{1, 2}, args)
}

func TestWherePartsAppendToSqlErr(t *testing.T) {
	parts := []Sqlizer{newWherePart(1)}
	_, err := appendToSql(parts, &bytes.Buffer{}, "", []interface{}{})
	assert.Error(t, err)
}

func TestWherePartNil(t *testing.T) {
	sql, _, _ := newWherePart(nil).ToSql()
	assert.Equal(t, "", sql)
}

func TestWherePartErr(t *testing.T) {
	_, _, err := newWherePart(1).ToSql()
	assert.Error(t, err)
}

func TestWherePartString(t *testing.T) {
	sql, args, _ := newWherePart("x = ?", 1).ToSql()
	assert.Equal(t, "x = ?", sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestWherePartMap(t *testing.T) {
	test := func(pred interface{}) {
		sql, _, _ := newWherePart(pred).ToSql()
		expect := []string{"x = ? AND y = ?", "y = ? AND x = ?"}
		if sql != expect[0] && sql != expect[1] {
			t.Errorf("expected one of %#v, got %#v", expect, sql)
		}
	}
	m := map[string]interface{}{"x": 1, "y": 2}
	test(m)
	test(Eq(m))
}

func TestWhereAutoParenthesisIssue380(t *testing.T) {
	sql, args, err := Select("*").
		From("table").
		Where("a = ?", 1).
		Where("b = ? OR c = ?", 2, 3).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM table WHERE a = ? AND (b = ? OR c = ?)", sql)
	assert.Equal(t, []interface{}{1, 2, 3}, args)
}

func TestWhereAutoParenthesisCaseInsensitive(t *testing.T) {
	cases := []struct {
		clause   string
		expected string
	}{
		{"b = ? OR c = ?", "(b = ? OR c = ?)"},
		{"b = ? or c = ?", "(b = ? or c = ?)"},
		{"b = ? Or c = ?", "(b = ? Or c = ?)"},
		{"b = ? oR c = ?", "(b = ? oR c = ?)"},
		{"a = ? OR b = ? OR c = ?", "(a = ? OR b = ? OR c = ?)"},
	}

	for _, tc := range cases {
		sql, _, err := Select("*").From("t").Where(tc.clause, 1, 2).ToSql()
		assert.NoError(t, err)
		assert.Equal(t, "SELECT * FROM t WHERE "+tc.expected, sql)
	}
}

func TestWhereAutoParenthesisAlreadyParenthesized(t *testing.T) {
	sql, _, err := Select("*").From("t").
		Where("a = ?", 1).
		Where("(b = ? OR c = ?)", 2, 3).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM t WHERE a = ? AND (b = ? OR c = ?)", sql)
}

func TestWhereAutoParenthesisTopLevelOrMultipleParens(t *testing.T) {
	sql, _, err := Select("*").From("t").
		Where("a = ?", 1).
		Where("(b = ?) OR (c = ?)", 2, 3).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM t WHERE a = ? AND ((b = ?) OR (c = ?))", sql)
}

func TestWhereAutoParenthesisNestedBooleans(t *testing.T) {
	// Nested OR inside parenthesized clause: top-level operator is AND
	sql, _, err := Select("*").From("t").
		Where("x = ?", 0).
		Where("a = ? AND (b = ? OR c = ?)", 1, 2, 3).
		ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM t WHERE x = ? AND a = ? AND (b = ? OR c = ?)", sql)

	// Top-level OR with nested AND
	sql, _, err = Select("*").From("t").
		Where("x = ?", 0).
		Where("a = ? OR (b = ? AND c = ?)", 1, 2, 3).
		ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM t WHERE x = ? AND (a = ? OR (b = ? AND c = ?))", sql)
}

func TestWhereAutoParenthesisSqlizer(t *testing.T) {
	sql, args, err := Select("*").From("t").
		Where("a = ?", 1).
		Where(Expr("b = ? OR c = ?", 2, 3)).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM t WHERE a = ? AND (b = ? OR c = ?)", sql)
	assert.Equal(t, []interface{}{1, 2, 3}, args)

	// sq.Or should not be double parenthesized
	sql, args, err = Select("*").From("t").
		Where("a = ?", 1).
		Where(Or{Expr("b = ?", 2), Expr("c = ?", 3)}).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM t WHERE a = ? AND (b = ? OR c = ?)", sql)
	assert.Equal(t, []interface{}{1, 2, 3}, args)
}

func TestWhereAutoParenthesisWordBoundaries(t *testing.T) {
	nonOrClauses := []string{
		"order_id = ?",
		"origin = ?",
		"color = ?",
		"for_sale = ?",
		"status = 'ORDERED'",
		"note = 'foo OR bar'",
		"note = 'it''s OR not'",
		`"order" = ?`,
		"`order` = ?",
		"a = 1 XOR b = 2",
		"-- comment with OR\na = 1",
		"/* comment with OR */ a = 1",
	}

	for _, clause := range nonOrClauses {
		sql, _, err := Select("*").From("t").Where(clause, 1).ToSql()
		assert.NoError(t, err)
		assert.Equal(t, "SELECT * FROM t WHERE "+clause, sql)
	}

	// Comments surrounding real OR clauses
	sql, _, err := Select("*").From("t").
		Where("a = 1 /* block comment */ OR b = 2").
		ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM t WHERE (a = 1 /* block comment */ OR b = 2)", sql)

	sql, _, err = Select("*").From("t").
		Where("a = 1 -- line comment\nOR b = 2").
		ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM t WHERE (a = 1 -- line comment\nOR b = 2)", sql)
}

func TestParenHelper(t *testing.T) {
	// String predicate
	sql, args, err := Paren("a = ? OR b = ?", 1, 2).ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "(a = ? OR b = ?)", sql)
	assert.Equal(t, []interface{}{1, 2}, args)

	// Map predicate
	sql, args, err = Paren(Eq{"a": 1, "b": 2}).ToSql()
	assert.NoError(t, err)
	expect := []string{"(a = ? AND b = ?)", "(b = ? AND a = ?)"}
	assert.True(t, sql == expect[0] || sql == expect[1])
	assert.Equal(t, 2, len(args))

	// Already enclosed predicate
	sql, args, err = Paren("(a = 1)").ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "(a = 1)", sql)
	assert.Empty(t, args)

	// ParenExpr helper
	sql, args, err = ParenExpr("x = ?", 1).ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "(x = ?)", sql)
	assert.Equal(t, []interface{}{1}, args)

	// Empty and nil
	sql, _, err = Paren(nil).ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "", sql)

	sql, _, err = Paren("").ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "", sql)

	// Invalid type
	_, _, err = Paren(123).ToSql()
	assert.Error(t, err)

	// Used inside builder Where
	sql, args, err = Select("*").From("t").
		Where("x = ?", 1).
		Where(Paren("y = ? AND z = ?", 2, 3)).
		ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM t WHERE x = ? AND (y = ? AND z = ?)", sql)
	assert.Equal(t, []interface{}{1, 2, 3}, args)
}

func TestDeleteBuilderAutoParenthesis(t *testing.T) {
	sql, args, err := Delete("users").
		Where("active = ?", false).
		Where("age < ? OR age > ?", 18, 65).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "DELETE FROM users WHERE active = ? AND (age < ? OR age > ?)", sql)
	assert.Equal(t, []interface{}{false, 18, 65}, args)
}

func TestUpdateBuilderAutoParenthesis(t *testing.T) {
	sql, args, err := Update("users").
		Set("status", "reviewed").
		Where("active = ?", true).
		Where("score < ? OR score > ?", 50, 90).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "UPDATE users SET status = ? WHERE active = ? AND (score < ? OR score > ?)", sql)
	assert.Equal(t, []interface{}{"reviewed", true, 50, 90}, args)
}

func TestSelectBuilderHavingAutoParenthesis(t *testing.T) {
	sql, args, err := Select("dept", "COUNT(*)").
		From("employees").
		GroupBy("dept").
		Having("COUNT(*) > ?", 5).
		Having("SUM(salary) < ? OR AVG(salary) > ?", 10000, 50000).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT dept, COUNT(*) FROM employees GROUP BY dept HAVING COUNT(*) > ? AND (SUM(salary) < ? OR AVG(salary) > ?)", sql)
	assert.Equal(t, []interface{}{5, 10000, 50000}, args)
}

func TestStatementBuilderAutoParenthesis(t *testing.T) {
	sb := StatementBuilder.Where("active = ?", true)
	sql, args, err := sb.Select("*").
		From("records").
		Where("flag1 = ? OR flag2 = ?", 1, 2).
		ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM records WHERE active = ? AND (flag1 = ? OR flag2 = ?)", sql)
	assert.Equal(t, []interface{}{true, 1, 2}, args)
}
