package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDropBuilderToSql(t *testing.T) {
	b := Drop("a").
		Prefix("WITH prefix AS ?", 0).
		Suffix("RETURNING ?", 4)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql :=
		"WITH prefix AS ? DROP TABLE a RETURNING ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0, 4}
	assert.Equal(t, expectedArgs, args)
}

func TestDropBuilderToSqlErr(t *testing.T) {
	_, _, err := Drop("").ToSql()
	assert.Error(t, err)
}

func TestDropBuilderMustSql(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestDropBuilderMustSql should have panicked!")
		}
	}()
	Drop("").MustSql()
}

func TestDropBuilderPlaceholders(t *testing.T) {
	b := Drop("test").Suffix("SUFFIX x = ? AND y = ?", 1, 2)

	sql, _, _ := b.PlaceholderFormat(Question).ToSql()
	assert.Equal(t, "DROP TABLE test SUFFIX x = ? AND y = ?", sql)

	sql, _, _ = b.PlaceholderFormat(Dollar).ToSql()
	assert.Equal(t, "DROP TABLE test SUFFIX x = $1 AND y = $2", sql)
}

func TestDropBuilderRunners(t *testing.T) {
	db := &DBStub{}
	b := Drop("test").Suffix("x = ?", 1).RunWith(db)

	expectedSql := "DROP TABLE test x = ?"

	b.Exec()
	assert.Equal(t, expectedSql, db.LastExecSql)
}

func TestDropBuilderNoRunner(t *testing.T) {
	b := Drop("test")

	_, err := b.Exec()
	assert.Equal(t, RunnerNotSet, err)
}

func TestDropWithQuery(t *testing.T) {
	db := &DBStub{}
	b := Drop("test").Suffix("RETURNING path").RunWith(db)

	expectedSql := "DROP TABLE test RETURNING path"
	b.Query()

	assert.Equal(t, expectedSql, db.LastQuerySql)
}
