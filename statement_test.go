package squirrel

import (
	"database/sql"
	"testing"

	"github.com/lann/builder"
	"github.com/stretchr/testify/assert"
)

func TestStatementBuilder(t *testing.T) {
	db := &DBStub{}
	sb := StatementBuilder.RunWith(db).SerializeWith(DefaultSerializer{})

	sb.Select("test").Exec()
	assert.Equal(t, "SELECT test", db.LastExecSql)
}

func TestStatementBuilderPlaceholderFormat(t *testing.T) {
	db := &DBStub{}
	sb := StatementBuilder.RunWith(db).SerializeWith(DefaultSerializer{}).PlaceholderFormat(Dollar)

	sb.Select("test").Where("x = ?").Exec()
	assert.Equal(t, "SELECT test WHERE x = $1", db.LastExecSql)
}

func TestRunWithDB(t *testing.T) {
	db := &sql.DB{}
	assert.NotPanics(t, func() {
		builder.GetStruct(Select().RunWith(db).SerializeWith(DefaultSerializer{}))
		builder.GetStruct(Insert("t").RunWith(db).SerializeWith(DefaultSerializer{}))
		builder.GetStruct(Update("t").RunWith(db).SerializeWith(DefaultSerializer{}))
		builder.GetStruct(Delete("t").RunWith(db).SerializeWith(DefaultSerializer{}))
	}, "RunWith(*sql.DB) should not panic")

}

func TestRunWithTx(t *testing.T) {
	tx := &sql.Tx{}
	assert.NotPanics(t, func() {
		builder.GetStruct(Select().RunWith(tx).SerializeWith(DefaultSerializer{}))
		builder.GetStruct(Insert("t").RunWith(tx).SerializeWith(DefaultSerializer{}))
		builder.GetStruct(Update("t").RunWith(tx).SerializeWith(DefaultSerializer{}))
		builder.GetStruct(Delete("t").RunWith(tx).SerializeWith(DefaultSerializer{}))
	}, "RunWith(*sql.Tx) should not panic")
}
