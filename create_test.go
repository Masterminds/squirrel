package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateBuilderToSql(t *testing.T) {
	b := Create("t").
		Column("id", "INT", "NOT NULL", "AUTOINCREMENT", "PRIMARY KEY").
		Column("c", "VARCHAR(32)")

	sql, _, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql := "CREATE TABLE t (" +
		"id INT NOT NULL AUTOINCREMENT PRIMARY KEY," +
		"c VARCHAR(32))"

	assert.Equal(t, expectedSql, sql)
}

func TestCreateBuilderToSqlErr(t *testing.T) {
	b := Create("t")
	_, _, err := b.ToSql()
	assert.Error(t, err)
}
