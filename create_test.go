package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateBuilderToSql(t *testing.T) {
	b := Create("test").
		Prefix("WITH prefix AS ?", 0).
		Columns("id", "a", "b").
		Types("INT", "TEXT", "DATE").
		PrimaryKey("id").
		Suffix("RETURNING ?", 4)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql :=
		"WITH prefix AS ? " +
			"CREATE TABLE test ( id INT, a TEXT, b DATE, PRIMARY KEY (id) ) " +
			"RETURNING ?"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0, 4}
	assert.Equal(t, expectedArgs, args)
}

func TestCreateBuilderSetMapToSql(t *testing.T) {
	b := Create("test").
		SetMap(map[string]string{
			"a": "INT",
			"b": "TEXT",
			"c": "DATE",
			"d": "TIMESTAMP",
		})

	sql, _, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "CREATE TABLE test ( a INT, b TEXT, c DATE, d TIMESTAMP )"
	assert.Equal(t, expectedSql, sql)
}

func TestCreateBuilderToSqlErr(t *testing.T) {
	_, _, err := Create("").ToSql()
	assert.Error(t, err)
}

func TestCreateBuilderMustSql(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestCreateBuilderMustSql should have panicked!")
		}
	}()
	Create("").MustSql()
}

func TestCreateBuilderPlaceholders(t *testing.T) {
	b := Create("test").Columns("id", "a", "b").
		Types("INT", "TEXT", "DATE").Suffix("SUFFIX x = ? AND y = ?")

	sql, _, _ := b.PlaceholderFormat(Question).ToSql()
	assert.Equal(t, "CREATE TABLE test ( id INT, a TEXT, b DATE ) SUFFIX x = ? AND y = ?", sql)

	sql, _, _ = b.PlaceholderFormat(Dollar).ToSql()
	assert.Equal(t, "CREATE TABLE test ( id INT, a TEXT, b DATE ) SUFFIX x = $1 AND y = $2", sql)
}

func TestCreateBuilderRunners(t *testing.T) {
	db := &DBStub{}
	b := Create("test").Columns("id", "a", "b").
		Types("INT", "TEXT", "DATE").Suffix("SUFFIX x = ?").RunWith(db)

	expectedSql := "CREATE TABLE test ( id INT, a TEXT, b DATE ) SUFFIX x = ?"

	b.Exec()
	assert.Equal(t, expectedSql, db.LastExecSql)
}

func TestCreateBuilderNoRunner(t *testing.T) {
	b := Create("test")

	_, err := b.Exec()
	assert.Equal(t, RunnerNotSet, err)
}

func TestCreateWithQuery(t *testing.T) {
	db := &DBStub{}
	b := Create("test").Columns("id", "a", "b").
		Types("INT", "TEXT", "DATE").Suffix("RETURNING path").RunWith(db)

	expectedSql := "CREATE TABLE test ( id INT, a TEXT, b DATE ) RETURNING path"
	b.Query()

	assert.Equal(t, expectedSql, db.LastQuerySql)
}
