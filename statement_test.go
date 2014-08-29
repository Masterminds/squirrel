package squirrel

import (
	"database/sql"
	"testing"

	"github.com/lann/builder"
	"github.com/stretchr/testify/assert"
)

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

func TestRunWithDB(t *testing.T) {
	db := &sql.DB{}
	assert.NotPanics(t, func() {
		builder.GetStruct(Select().RunWith(db))
		builder.GetStruct(Insert("t").RunWith(db))
		builder.GetStruct(Update("t").RunWith(db))
		builder.GetStruct(Delete("t").RunWith(db))
	}, "RunWith(*sql.DB) should not panic")

}

func TestRunWithTx(t *testing.T) {
	tx := &sql.Tx{}
	assert.NotPanics(t, func() {
		builder.GetStruct(Select().RunWith(tx))
		builder.GetStruct(Insert("t").RunWith(tx))
		builder.GetStruct(Update("t").RunWith(tx))
		builder.GetStruct(Delete("t").RunWith(tx))
	}, "RunWith(*sql.Tx) should not panic")
}
