package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteBuilderToSql(t *testing.T) {
	b := Delete("").
		Prefix("WITH prefix AS ?", 0).
		From("a").
		Using("u2").
		JoinClause("CROSS JOIN j1").
		Join("j2").
		LeftJoin("j3").
		RightJoin("j4").
		InnerJoin("j5").
		CrossJoin("j6").
		Where("b = ?", 1).
		OrderBy("c").
		Limit(2).
		Offset(3).
		Suffix("RETURNING ?", 4)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql :=
		"WITH prefix AS ? " +
			"DELETE FROM a USING u2 " +
			"CROSS JOIN j1 JOIN j2 LEFT JOIN j3 RIGHT JOIN j4 INNER JOIN j5 CROSS JOIN j6 " +
			"WHERE b = ? ORDER BY c LIMIT 2 OFFSET 3 " +
			"RETURNING ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0, 1, 4}
	assert.Equal(t, expectedArgs, args)
}

func TestDeleteBuilderToSqlErr(t *testing.T) {
	_, _, err := Delete("").ToSql()
	assert.Error(t, err)
}

func TestDeleteBuilderMustSql(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestDeleteBuilderMustSql should have panicked!")
		}
	}()
	Delete("").MustSql()
}

func TestDeleteBuilderPlaceholders(t *testing.T) {
	b := Delete("test").Where("x = ? AND y = ?", 1, 2)

	sql, _, _ := b.PlaceholderFormat(Question).ToSql()
	assert.Equal(t, "DELETE FROM test WHERE x = ? AND y = ?", sql)

	sql, _, _ = b.PlaceholderFormat(Dollar).ToSql()
	assert.Equal(t, "DELETE FROM test WHERE x = $1 AND y = $2", sql)
}

func TestDeleteBuilderRunners(t *testing.T) {
	db := &DBStub{}
	b := Delete("test").Where("x = ?", 1).RunWith(db)

	expectedSql := "DELETE FROM test WHERE x = ?"

	b.Exec()
	assert.Equal(t, expectedSql, db.LastExecSql)
}

func TestDeleteBuilderNoRunner(t *testing.T) {
	b := Delete("test")

	_, err := b.Exec()
	assert.Equal(t, RunnerNotSet, err)
}

func TestDeleteWithQuery(t *testing.T) {
	db := &DBStub{}
	b := Delete("test").Where("id=55").Suffix("RETURNING path").RunWith(db)

	expectedSql := "DELETE FROM test WHERE id=55 RETURNING path"
	b.Query()

	assert.Equal(t, expectedSql, db.LastQuerySql)
}

func TestDeleteWithMultipleUsings(t *testing.T) {
	b := Delete("a").
		Using("t1", "t2").
		Where(And{
			Eq{"a.t1_c1": "t1.c1"},
			Eq{"a.t2_c2": "t2.c2"},
			Eq{"a.id": int64(16)},
		}).
		PlaceholderFormat(Dollar)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql :=
		"DELETE FROM a " +
			"USING t1, t2 " +
			"WHERE (a.t1_c1 = $1 AND a.t2_c2 = $2 AND a.id = $3)"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{"t1.c1", "t2.c2", int64(16)}
	assert.Equal(t, expectedArgs, args)
}
