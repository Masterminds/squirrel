package squirrel

import (
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSelectBuilderToSql(t *testing.T) {
	subQ := Select("aa", "bb").From("dd")
	b := Select("a", "b").
		Prefix("WITH prefix AS ?", 0).
		Distinct().
		Columns("c").
		Column("IF(d IN ("+Placeholders(3)+"), 1, 0) as stat_column", 1, 2, 3).
		Column(Expr("a > ?", 100)).
		Column(Alias(Eq{"b": []int{101, 102, 103}}, "b_alias")).
		Column(Alias(subQ, "subq")).
		From("e").
		JoinClause("CROSS JOIN j1").
		Join("j2").
		LeftJoin("j3").
		RightJoin("j4").
		InnerJoin("j5").
		CrossJoin("j6").
		Where("f = ?", 4).
		Where(Eq{"g": 5}).
		Where(map[string]interface{}{"h": 6}).
		Where(Eq{"i": []int{7, 8, 9}}).
		Where(Or{Expr("j = ?", 10), And{Eq{"k": 11}, Expr("true")}}).
		GroupBy("l").
		Having("m = n").
		OrderByClause("? DESC", 1).
		OrderBy("o ASC", "p DESC").
		Limit(12).
		Offset(13).
		Suffix("FETCH FIRST ? ROWS ONLY", 14)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql :=
		"WITH prefix AS ? " +
			"SELECT DISTINCT a, b, c, IF(d IN (?,?,?), 1, 0) as stat_column, a > ?, " +
			"(b IN (?,?,?)) AS b_alias, " +
			"(SELECT aa, bb FROM dd) AS subq " +
			"FROM e " +
			"CROSS JOIN j1 JOIN j2 LEFT JOIN j3 RIGHT JOIN j4 INNER JOIN j5 CROSS JOIN j6 " +
			"WHERE f = ? AND g = ? AND h = ? AND i IN (?,?,?) AND (j = ? OR (k = ? AND true)) " +
			"GROUP BY l HAVING m = n ORDER BY ? DESC, o ASC, p DESC LIMIT 12 OFFSET 13 " +
			"FETCH FIRST ? ROWS ONLY"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0, 1, 2, 3, 100, 101, 102, 103, 4, 5, 6, 7, 8, 9, 10, 11, 1, 14}
	assert.Equal(t, expectedArgs, args)
}

func TestSelectBuilderFromSelect(t *testing.T) {
	subQ := Select("c").From("d").Where(Eq{"i": 0})
	b := Select("a", "b").FromSelect(subQ, "subq")
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT a, b FROM (SELECT c FROM d WHERE i = ?) AS subq"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0}
	assert.Equal(t, expectedArgs, args)
}

func TestSelectBuilderFromSelectNestedDollarPlaceholders(t *testing.T) {
	subQ := Select("c").
		From("t").
		Where(Gt{"c": 1}).
		PlaceholderFormat(Dollar)
	b := Select("c").
		FromSelect(subQ, "subq").
		Where(Lt{"c": 2}).
		PlaceholderFormat(Dollar)
	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT c FROM (SELECT c FROM t WHERE c > $1) AS subq WHERE c < $2"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1, 2}
	assert.Equal(t, expectedArgs, args)
}

func TestSelectBuilderToSqlErr(t *testing.T) {
	_, _, err := Select().From("x").ToSql()
	assert.Error(t, err)
}

func TestSelectBuilderPlaceholders(t *testing.T) {
	b := Select("test").Where("x = ? AND y = ?")

	sql, _, _ := b.PlaceholderFormat(Question).ToSql()
	assert.Equal(t, "SELECT test WHERE x = ? AND y = ?", sql)

	sql, _, _ = b.PlaceholderFormat(Dollar).ToSql()
	assert.Equal(t, "SELECT test WHERE x = $1 AND y = $2", sql)

	sql, _, _ = b.PlaceholderFormat(Colon).ToSql()
	assert.Equal(t, "SELECT test WHERE x = :1 AND y = :2", sql)

	sql, _, _ = b.PlaceholderFormat(AtP).ToSql()
	assert.Equal(t, "SELECT test WHERE x = @p1 AND y = @p2", sql)
}

func TestSelectBuilderRunners(t *testing.T) {
	db := &DBStub{}
	b := Select("test").RunWith(db)

	expectedSql := "SELECT test"

	b.Exec()
	assert.Equal(t, expectedSql, db.LastExecSql)

	b.Query()
	assert.Equal(t, expectedSql, db.LastQuerySql)

	b.QueryRow()
	assert.Equal(t, expectedSql, db.LastQueryRowSql)

	err := b.Scan()
	assert.NoError(t, err)
}

func TestSelectBuilderNoRunner(t *testing.T) {
	b := Select("test")

	_, err := b.Exec()
	assert.Equal(t, RunnerNotSet, err)

	_, err = b.Query()
	assert.Equal(t, RunnerNotSet, err)

	err = b.Scan()
	assert.Equal(t, RunnerNotSet, err)
}

func TestSelectBuilderSimpleJoin(t *testing.T) {

	expectedSql := "SELECT * FROM bar JOIN baz ON bar.foo = baz.foo"
	expectedArgs := []interface{}(nil)

	b := Select("*").From("bar").Join("baz ON bar.foo = baz.foo")

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	assert.Equal(t, expectedSql, sql)
	assert.Equal(t, args, expectedArgs)
}

func TestSelectBuilderParamJoin(t *testing.T) {

	expectedSql := "SELECT * FROM bar JOIN baz ON bar.foo = baz.foo AND baz.foo = ?"
	expectedArgs := []interface{}{42}

	b := Select("*").From("bar").Join("baz ON bar.foo = baz.foo AND baz.foo = ?", 42)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	assert.Equal(t, expectedSql, sql)
	assert.Equal(t, args, expectedArgs)
}

func TestSelectBuilderNestedSelectJoin(t *testing.T) {

	expectedSql := "SELECT * FROM bar JOIN ( SELECT * FROM baz WHERE foo = ? ) r ON bar.foo = r.foo"
	expectedArgs := []interface{}{42}

	nestedSelect := Select("*").From("baz").Where("foo = ?", 42)

	b := Select("*").From("bar").JoinClause(nestedSelect.Prefix("JOIN (").Suffix(") r ON bar.foo = r.foo"))

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	assert.Equal(t, expectedSql, sql)
	assert.Equal(t, args, expectedArgs)
}

func TestSelectWithOptions(t *testing.T) {
	sql, _, err := Select("*").From("foo").Distinct().Options("SQL_NO_CACHE").ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT DISTINCT SQL_NO_CACHE * FROM foo", sql)
}

func TestSelectWithRemoveLimit(t *testing.T) {
	sql, _, err := Select("*").From("foo").Limit(10).RemoveLimit().ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM foo", sql)
}

func TestSelectWithRemoveOffset(t *testing.T) {
	sql, _, err := Select("*").From("foo").Offset(10).RemoveOffset().ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM foo", sql)
}

func TestSelectBuilderNestedSelectDollar(t *testing.T) {
	nestedBuilder := StatementBuilder.PlaceholderFormat(Dollar).Select("*").Prefix("NOT EXISTS (").
		From("bar").Where("y = ?", 42).Suffix(")")
	outerSql, _, err := StatementBuilder.PlaceholderFormat(Dollar).Select("*").
		From("foo").Where("x = ?").Where(nestedBuilder).ToSql()

	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM foo WHERE x = $1 AND NOT EXISTS ( SELECT * FROM bar WHERE y = $2 )", outerSql)
}

func TestSelectBuilderMustSql(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestSelectBuilderMustSql should have panicked!")
		}
	}()
	// This function should cause a panic
	Select().From("foo").MustSql()
}

func TestSelectWithoutWhereClause(t *testing.T) {
	sql, _, err := Select("*").From("users").ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users", sql)
}

func TestSelectWithNilWhereClause(t *testing.T) {
	sql, _, err := Select("*").From("users").Where(nil).ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users", sql)
}

func TestSelectWithEmptyStringWhereClause(t *testing.T) {
	sql, _, err := Select("*").From("users").Where("").ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT * FROM users", sql)
}

func TestSelectSubqueryPlaceholderNumbering(t *testing.T) {
	subquery := Select("a").Where("b = ?", 1).PlaceholderFormat(Dollar)
	with := subquery.Prefix("WITH a AS (").Suffix(")")

	sql, args, err := Select("*").
		PrefixExpr(with).
		FromSelect(subquery, "q").
		Where("c = ?", 2).
		PlaceholderFormat(Dollar).
		ToSql()
	assert.NoError(t, err)

	expectedSql := "WITH a AS ( SELECT a WHERE b = $1 ) SELECT * FROM (SELECT a WHERE b = $2) AS q WHERE c = $3"
	assert.Equal(t, expectedSql, sql)
	assert.Equal(t, []interface{}{1, 1, 2}, args)
}

func TestSelectSubqueryInConjunctionPlaceholderNumbering(t *testing.T) {
	subquery := Select("a").Where(Eq{"b": 1}).Prefix("EXISTS(").Suffix(")").PlaceholderFormat(Dollar)

	sql, args, err := Select("*").
		Where(Or{subquery}).
		Where("c = ?", 2).
		PlaceholderFormat(Dollar).
		ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT * WHERE (EXISTS( SELECT a WHERE b = $1 )) AND c = $2"
	assert.Equal(t, expectedSql, sql)
	assert.Equal(t, []interface{}{1, 2}, args)
}

func TestSelectJoinClausePlaceholderNumbering(t *testing.T) {
	subquery := Select("a").Where(Eq{"b": 2}).PlaceholderFormat(Dollar)

	sql, args, err := Select("t1.a").
		From("t1").
		Where(Eq{"a": 1}).
		JoinClause(subquery.Prefix("JOIN (").Suffix(") t2 ON (t1.a = t2.a)")).
		PlaceholderFormat(Dollar).
		ToSql()
	assert.NoError(t, err)

	expectedSql := "SELECT t1.a FROM t1 JOIN ( SELECT a WHERE b = $1 ) t2 ON (t1.a = t2.a) WHERE a = $2"
	assert.Equal(t, expectedSql, sql)
	assert.Equal(t, []interface{}{2, 1}, args)
}

func ExampleSelect() {
	Select("id", "created", "first_name").From("users") // ... continue building up your query

	// sql methods in select columns are ok
	Select("first_name", "count(*)").From("users")

	// column aliases are ok too
	Select("first_name", "count(*) as n_users").From("users")
}

func ExampleSelectBuilder_From() {
	Select("id", "created", "first_name").From("users") // ... continue building up your query
}

func ExampleSelectBuilder_Where() {
	companyId := 20
	Select("id", "created", "first_name").From("users").Where("company = ?", companyId)
}

func ExampleSelectBuilder_Where_helpers() {
	companyId := 20

	Select("id", "created", "first_name").From("users").Where(Eq{
		"company": companyId,
	})

	Select("id", "created", "first_name").From("users").Where(GtOrEq{
		"created": time.Now().AddDate(0, 0, -7),
	})

	Select("id", "created", "first_name").From("users").Where(And{
		GtOrEq{
			"created": time.Now().AddDate(0, 0, -7),
		},
		Eq{
			"company": companyId,
		},
	})
}

func ExampleSelectBuilder_Where_multiple() {
	companyId := 20

	// multiple where's are ok

	Select("id", "created", "first_name").
		From("users").
		Where("company = ?", companyId).
		Where(GtOrEq{
			"created": time.Now().AddDate(0, 0, -7),
		})
}

func ExampleSelectBuilder_FromSelect() {
	usersByCompany := Select("company", "count(*) as n_users").From("users").GroupBy("company")
	query := Select("company.id", "company.name", "users_by_company.n_users").
		FromSelect(usersByCompany, "users_by_company").
		Join("company on company.id = users_by_company.company")

	sql, _, _ := query.ToSql()
	fmt.Println(sql)

	// Output: SELECT company.id, company.name, users_by_company.n_users FROM (SELECT company, count(*) as n_users FROM users GROUP BY company) AS users_by_company JOIN company on company.id = users_by_company.company
}

func ExampleSelectBuilder_Columns() {
	query := Select("id").Columns("created", "first_name").From("users")

	sql, _, _ := query.ToSql()
	fmt.Println(sql)
	// Output: SELECT id, created, first_name FROM users
}

func ExampleSelectBuilder_Columns_order() {
	// out of order is ok too
	query := Select("id").Columns("created").From("users").Columns("first_name")

	sql, _, _ := query.ToSql()
	fmt.Println(sql)
	// Output: SELECT id, created, first_name FROM users
}

func ExampleSelectBuilder_Scan() {

	var db *sql.DB

	query := Select("id", "created", "first_name").From("users")
	query = query.RunWith(db)

	var id int
	var created time.Time
	var firstName string

	if err := query.Scan(&id, &created, &firstName); err != nil {
		log.Println(err)
		return
	}
}

func ExampleSelectBuilder_ScanContext() {

	var db *sql.DB

	query := Select("id", "created", "first_name").From("users")
	query = query.RunWith(db)

	var id int
	var created time.Time
	var firstName string

	if err := query.ScanContext(ctx, &id, &created, &firstName); err != nil {
		log.Println(err)
		return
	}
}

func ExampleSelectBuilder_RunWith() {

	var db *sql.DB

	query := Select("id", "created", "first_name").From("users").RunWith(db)

	var id int
	var created time.Time
	var firstName string

	if err := query.Scan(&id, &created, &firstName); err != nil {
		log.Println(err)
		return
	}
}

func ExampleSelectBuilder_ToSql() {

	var db *sql.DB

	query := Select("id", "created", "first_name").From("users")

	sql, args, err := query.ToSql()
	if err != nil {
		log.Println(err)
		return
	}

	rows, err := db.Query(sql, args...)
	if err != nil {
		log.Println(err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		// scan...
	}
}

func TestRemoveColumns(t *testing.T) {
	query := Select("id").
		From("users").
		RemoveColumns()
	query = query.Columns("name")
	sql, _, err := query.ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT name FROM users", sql)
}
