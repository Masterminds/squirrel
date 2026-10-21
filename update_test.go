package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateBuilderToSql(t *testing.T) {
	b := Update("").
		Prefix("WITH prefix AS ?", 0).
		Table("a").
		Set("b", Expr("? + 1", 1)).
		SetMap(Eq{"c": 2}).
		Set("c1", Case("status").When("1", "2").When("2", "1")).
		Set("c2", Case().When("a = 2", Expr("?", "foo")).When("a = 3", Expr("?", "bar"))).
		Set("c3", Select("a").From("b")).
		Where("d = ?", 3).
		OrderBy("e").
		Limit(4).
		Offset(5).
		Suffix("RETURNING ?", 6)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql :=
		"WITH prefix AS ? " +
			"UPDATE a SET b = ? + 1, c = ?, " +
			"c1 = CASE status WHEN 1 THEN 2 WHEN 2 THEN 1 END, " +
			"c2 = CASE WHEN a = 2 THEN ? WHEN a = 3 THEN ? END, " +
			"c3 = (SELECT a FROM b) " +
			"WHERE d = ? " +
			"ORDER BY e LIMIT 4 OFFSET 5 " +
			"RETURNING ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0, 1, 2, "foo", "bar", 3, 6}
	assert.Equal(t, expectedArgs, args)
}

func TestUpdateBuilderToSqlErr(t *testing.T) {
	_, _, err := Update("").Set("x", 1).ToSql()
	assert.Error(t, err)

	_, _, err = Update("x").ToSql()
	assert.Error(t, err)
}

func TestUpdateBuilderMustSql(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestUpdateBuilderMustSql should have panicked!")
		}
	}()
	Update("").MustSql()
}

func TestUpdateBuilderPlaceholders(t *testing.T) {
	b := Update("test").SetMap(Eq{"x": 1, "y": 2})

	sql, _, _ := b.PlaceholderFormat(Question).ToSql()
	assert.Equal(t, "UPDATE test SET x = ?, y = ?", sql)

	sql, _, _ = b.PlaceholderFormat(Dollar).ToSql()
	assert.Equal(t, "UPDATE test SET x = $1, y = $2", sql)
}

func TestUpdateBuilderRunners(t *testing.T) {
	db := &DBStub{}
	b := Update("test").Set("x", 1).RunWith(db)

	expectedSql := "UPDATE test SET x = ?"

	b.Exec()
	assert.Equal(t, expectedSql, db.LastExecSql)
}

func TestUpdateBuilderNoRunner(t *testing.T) {
	b := Update("test").Set("x", 1)

	_, err := b.Exec()
	assert.Equal(t, RunnerNotSet, err)
}

func TestUpdateBuilderFrom(t *testing.T) {
	sql, _, err := Update("employees").Set("sales_count", 100).From("accounts").Where("accounts.name = ?", "ACME").ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "UPDATE employees SET sales_count = ? FROM accounts WHERE accounts.name = ?", sql)
}

func TestUpdateBuilderFromSelect(t *testing.T) {
	sql, _, err := Update("employees").
		Set("sales_count", 100).
		FromSelect(Select("id").
			From("accounts").
			Where("accounts.name = ?", "ACME"), "subquery").
		Where("employees.account_id = subquery.id").ToSql()
	assert.NoError(t, err)

	expectedSql :=
		"UPDATE employees " +
			"SET sales_count = ? " +
			"FROM (SELECT id FROM accounts WHERE accounts.name = ?) AS subquery " +
			"WHERE employees.account_id = subquery.id"
	assert.Equal(t, expectedSql, sql)
}

func TestUpdateBuilderSubqueryPlaceholderFormats(t *testing.T) {
	testCases := []struct {
		name              string
		placeholderFormat PlaceholderFormat
		expectedSQL       string
	}{
		{"Dollar", Dollar, "UPDATE table1 SET initial_count = $1, count_column = (SELECT count(*) AS count FROM table0 WHERE arg1 = $2) WHERE arg2 = $3"},
		{"Question", Question, "UPDATE table1 SET initial_count = ?, count_column = (SELECT count(*) AS count FROM table0 WHERE arg1 = ?) WHERE arg2 = ?"},
		{"Colon", Colon, "UPDATE table1 SET initial_count = :1, count_column = (SELECT count(*) AS count FROM table0 WHERE arg1 = :2) WHERE arg2 = :3"},
		{"AtP", AtP, "UPDATE table1 SET initial_count = @p1, count_column = (SELECT count(*) AS count FROM table0 WHERE arg1 = @p2) WHERE arg2 = @p3"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			psql := StatementBuilder.PlaceholderFormat(tc.placeholderFormat)
			subQuery := psql.Select("count(*) AS count").From("table0").Where(Eq{"arg1": "inner"})
			sql, args, err := psql.Update("table1").
				Set("initial_count", "set").
				Set("count_column", subQuery).
				Where(Eq{"arg2": "outer"}).ToSql()
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, []interface{}{"set", "inner", "outer"}, args)
		})
	}
}
