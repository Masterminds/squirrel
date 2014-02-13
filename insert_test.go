package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertBuilderToSql(t *testing.T) {
	b := Insert("").
		Into("a").
		Columns("b", "c").
		Values(1, 2).
		Values(3, Expr("? + 1", 4))

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql :=
		"INSERT INTO a (b,c) VALUES (?,?),(?,? + 1)"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{1, 2, 3, 4}
	assert.Equal(t, expectedArgs, args)
}

func TestInsertBuilderToSqlErr(t *testing.T) {
	_, _, err := Insert("").Values(1).ToSql()
	assert.Error(t, err)

	_, _, err = Insert("x").ToSql()
	assert.Error(t, err)
}

func TestInsertBuilderPlaceholders(t *testing.T) {
	b := Insert("test").Values(1, 2)

	sql, _, _ := b.PlaceholderFormat(Question).ToSql()
	assert.Equal(t, "INSERT INTO test VALUES (?,?)", sql)

	sql, _, _ = b.PlaceholderFormat(Dollar).ToSql()
	assert.Equal(t, "INSERT INTO test VALUES ($1,$2)", sql)
}

func TestInsertBuilderRunners(t *testing.T) {
	db := &DBStub{}
	b := Insert("test").Values(1).RunWith(db)

	expectedSql := "INSERT INTO test VALUES (?)"

	b.Exec()
	assert.Equal(t, expectedSql, db.LastExecSql)
}

func TestInsertBuilderNoRunner(t *testing.T) {
	b := Insert("test").Values(1)

	_, err := b.Exec()
	assert.Equal(t, RunnerNotSet, err)
}
