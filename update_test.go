package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateBuilderToSQL(t *testing.T) {
	b := Update("").
		Prefix("WITH prefix AS ?", 0).
		Table("a").
		Set("b", Expr("? + 1", 1)).
		SetMap(Eq{"c": 2}).
		Where("d = ?", 3).
		OrderBy("e").
		Limit(4).
		Offset(5).
		Suffix("RETURNING ?", 6)

	sql, args, err := b.ToSQL()
	assert.NoError(t, err)

	expectedSQL :=
		"WITH prefix AS ? " +
			"UPDATE a SET b = ? + 1, c = ? WHERE d = ? " +
			"ORDER BY e LIMIT 4 OFFSET 5 " +
			"RETURNING ?"
	assert.Equal(t, expectedSQL, sql)

	expectedArgs := []interface{}{0, 1, 2, 3, 6}
	assert.Equal(t, expectedArgs, args)
}

func TestUpdateBuilderToSQLErr(t *testing.T) {
	_, _, err := Update("").Set("x", 1).ToSQL()
	assert.Error(t, err)

	_, _, err = Update("x").ToSQL()
	assert.Error(t, err)
}

func TestUpdateBuilderPlaceholders(t *testing.T) {
	b := Update("test").SetMap(Eq{"x": 1, "y": 2})

	sql, _, _ := b.PlaceholderFormat(Question).ToSQL()
	assert.Equal(t, "UPDATE test SET x = ?, y = ?", sql)

	sql, _, _ = b.PlaceholderFormat(Dollar).ToSQL()
	assert.Equal(t, "UPDATE test SET x = $1, y = $2", sql)
}

func TestUpdateBuilderRunners(t *testing.T) {
	db := &DBStub{}
	b := Update("test").Set("x", 1).RunWith(db)

	expectedSQL := "UPDATE test SET x = ?"

	b.Exec()
	assert.Equal(t, expectedSQL, db.LastExecSQL)
}

func TestUpdateBuilderNoRunner(t *testing.T) {
	b := Update("test").Set("x", 1)

	_, err := b.Exec()
	assert.Equal(t, RunnerNotSet, err)
}
