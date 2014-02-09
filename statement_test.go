package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelect(t *testing.T) {
	sql, _, _ := Select("test").ToSql()
	assert.Equal(t, "SELECT test", sql)
}

func TestStatementBuilder(t *testing.T) {
	db := &DBStub{}
	sb := StatementBuilder.RunWith(db)

	sb.Select("test").Exec()
	assert.Equal(t, "SELECT test", db.LastExecSql)
}

func TestStatementBuilderPlaceholderFormat(t *testing.T) {
	db := &DBStub{}
	sb := StatementBuilder.RunWith(db).PlaceholderFormat(Dollar)

	sb.Select("test").Where("x = ?").Exec()
	assert.Equal(t, "SELECT test WHERE x = $1", db.LastExecSql)
}
