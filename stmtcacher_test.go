package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStmtCacherPrepare(t *testing.T) {
	db := &DBStub{}
	sc := NewStmtCacher(db)
	query := "SELECT 1"

	sc.Prepare(query)
	assert.Equal(t, query, db.LastPrepareSql)

	sc.Prepare(query)
	assert.Equal(t, 1, db.PrepareCount, "expected 1 Prepare, got %d", db.PrepareCount)

	assert.Nil(t, sc.Close())

	_, err := sc.Prepare(query)
	assert.Error(t, err, "statement cache closed")
	assert.Equal(t, 1, db.PrepareCount, "expected 1 Prepare, got %d", db.PrepareCount)
}
