package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaseWithVal(t *testing.T) {
	caseStmt := Case("number").
		When("1", "one").
		When("2", "two").
		Else(Expr("?", "big number"))

	qb := Select().
		Column(caseStmt).
		From("table")
	sql, args, err := qb.ToSql()

	assert.NoError(t, err)

	expectedSql := "SELECT CASE number " +
		"WHEN 1 THEN one " +
		"WHEN 2 THEN two " +
		"ELSE ? " +
		"END " +
		"FROM table"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{"big number"}
	assert.Equal(t, expectedArgs, args)
}

func TestCaseWithComplexVal(t *testing.T) {
	caseStmt := Case("? > ?", 10, 5).
		When("true", "'T'")

	qb := Select().
		Column(Alias(caseStmt, "complexCase")).
		From("table")
	sql, args, err := qb.ToSql()

	assert.NoError(t, err)

	expectedSql := "SELECT (CASE ? > ? " +
		"WHEN true THEN 'T' " +
		"END) AS complexCase " +
		"FROM table"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{10, 5}
	assert.Equal(t, expectedArgs, args)
}

func TestCaseWithNoVal(t *testing.T) {
	caseStmt := Case().
		When(Eq{"x": 0}, "x is zero").
		When(Expr("x > ?", 1), Expr("CONCAT('x is greater than ', ?)", 2))

	qb := Select().Column(caseStmt).From("table")
	sql, args, err := qb.ToSql()

	assert.NoError(t, err)

	expectedSql := "SELECT CASE " +
		"WHEN x = ? THEN x is zero " +
		"WHEN x > ? THEN CONCAT('x is greater than ', ?) " +
		"END " +
		"FROM table"

	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0, 1, 2}
	assert.Equal(t, expectedArgs, args)
}

func TestCaseWithExpr(t *testing.T) {
	caseStmt := Case(Expr("x = ?", true)).
		When("true", Expr("?", "it's true!")).
		Else("42")

	qb := Select().Column(caseStmt).From("table")
	sql, args, err := qb.ToSql()

	assert.NoError(t, err)

	expectedSql := "SELECT CASE x = ? " +
		"WHEN true THEN ? " +
		"ELSE 42 " +
		"END " +
		"FROM table"

	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{true, "it's true!"}
	assert.Equal(t, expectedArgs, args)
}

func TestMultipleCase(t *testing.T) {
	caseStmtNoval := Case(Expr("x = ?", true)).
		When("true", Expr("?", "it's true!")).
		Else("42")
	caseStmtExpr := Case().
		When(Eq{"x": 0}, "'x is zero'").
		When(Expr("x > ?", 1), Expr("CONCAT('x is greater than ', ?)", 2))

	qb := Select().
		Column(Alias(caseStmtNoval, "case_noval")).
		Column(Alias(caseStmtExpr, "case_expr")).
		From("table")

	sql, args, err := qb.ToSql()

	assert.NoError(t, err)

	expectedSql := "SELECT " +
		"(CASE x = ? WHEN true THEN ? ELSE 42 END) AS case_noval, " +
		"(CASE WHEN x = ? THEN 'x is zero' WHEN x > ? THEN CONCAT('x is greater than ', ?) END) AS case_expr " +
		"FROM table"

	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{
		true, "it's true!",
		0, 1, 2,
	}
	assert.Equal(t, expectedArgs, args)
}

func TestCaseWithNoWhenClause(t *testing.T) {
	caseStmt := Case("something").
		Else("42")

	qb := Select().Column(caseStmt).From("table")

	_, _, err := qb.ToSql()

	assert.Error(t, err)

	assert.Equal(t, "case expression must contain at lease one WHEN clause", err.Error())
}

func TestCaseBuilderMustSql(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestCaseBuilderMustSql should have panicked!")
		}
	}()
	Case("").MustSql()
}

func TestCaseWithIntegerValues(t *testing.T) {
	caseStmt := Case("order_no").
		When("ORD001", 500).
		When("ORD002", 600)

	sql, args, err := Update("orders").
		Set("amount", caseStmt).
		Where(Eq{"order_no": []string{"ORD001", "ORD002", "ORD003"}}).
		ToSql()

	assert.NoError(t, err)

	expectedSql := "UPDATE orders SET amount = CASE order_no WHEN ORD001 THEN ? WHEN ORD002 THEN ? END " +
		"WHERE order_no IN (?,?,?)"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{500, 600, "ORD001", "ORD002", "ORD003"}
	assert.Equal(t, expectedArgs, args)
}

func TestCaseWithPrimitivesInWhenAndElse(t *testing.T) {
	caseStmt := Case("status").
		When(1, 100).
		When(2, 200).
		Else(0)

	qb := Select().Column(caseStmt).From("table")
	sql, args, err := qb.ToSql()

	assert.NoError(t, err)

	expectedSql := "SELECT CASE status " +
		"WHEN ? THEN ? " +
		"WHEN ? THEN ? " +
		"ELSE ? " +
		"END " +
		"FROM table"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1, 100, 2, 200, 0}
	assert.Equal(t, expectedArgs, args)
}

func TestCaseWithPrimitivesInWhat(t *testing.T) {
	caseStmt := Case(1).
		When(1, "one").
		When(2, "two").
		Else("other")

	qb := Select().Column(caseStmt).From("table")
	sql, args, err := qb.ToSql()

	assert.NoError(t, err)

	expectedSql := "SELECT CASE ? " +
		"WHEN ? THEN one " +
		"WHEN ? THEN two " +
		"ELSE other " +
		"END " +
		"FROM table"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1, 1, 2}
	assert.Equal(t, expectedArgs, args)
}

func TestCaseWithFloatBoolAndUint(t *testing.T) {
	caseStmt := Case(true).
		When(true, 1.25).
		When(false, 2.5).
		Else(uint(0))

	qb := Select().Column(caseStmt).From("table")
	sql, args, err := qb.ToSql()

	assert.NoError(t, err)

	expectedSql := "SELECT CASE ? " +
		"WHEN ? THEN ? " +
		"WHEN ? THEN ? " +
		"ELSE ? " +
		"END " +
		"FROM table"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{true, true, 1.25, false, 2.5, uint(0)}
	assert.Equal(t, expectedArgs, args)
}

func TestCaseWithMapCondition(t *testing.T) {
	caseStmt := Case().
		When(map[string]interface{}{"status": "pending"}, 10).
		Else(0)

	qb := Select().Column(caseStmt).From("table")
	sql, args, err := qb.ToSql()

	assert.NoError(t, err)

	expectedSql := "SELECT CASE " +
		"WHEN status = ? THEN ? " +
		"ELSE ? " +
		"END " +
		"FROM table"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{"pending", 10, 0}
	assert.Equal(t, expectedArgs, args)
}

func TestCaseWithWhatMethod(t *testing.T) {
	caseStmt := Case().
		What("category").
		When(1, 10).
		Else(0)

	qb := Select().Column(caseStmt).From("table")
	sql, args, err := qb.ToSql()

	assert.NoError(t, err)

	expectedSql := "SELECT CASE category " +
		"WHEN ? THEN ? " +
		"ELSE ? " +
		"END " +
		"FROM table"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1, 10, 0}
	assert.Equal(t, expectedArgs, args)
}

func TestCaseWithDollarPlaceholders(t *testing.T) {
	caseStmt := Case("order_no").
		When("ORD001", 500).
		When("ORD002", 600).
		Else(0)

	qb := StatementBuilder.PlaceholderFormat(Dollar).
		Select().
		Column(caseStmt).
		From("orders")
	sql, args, err := qb.ToSql()

	assert.NoError(t, err)

	expectedSql := "SELECT CASE order_no " +
		"WHEN ORD001 THEN $1 " +
		"WHEN ORD002 THEN $2 " +
		"ELSE $3 " +
		"END " +
		"FROM orders"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{500, 600, 0}
	assert.Equal(t, expectedArgs, args)
}

func TestCaseWithNilValues(t *testing.T) {
	caseStmt := Case("status").
		When(1, nil).
		Else(nil)

	sql, args, err := caseStmt.ToSql()

	assert.NoError(t, err)

	expectedSql := "CASE status WHEN ? THEN ? ELSE ? END"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1, nil, nil}
	assert.Equal(t, expectedArgs, args)
}
