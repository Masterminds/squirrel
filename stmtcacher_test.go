package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStmtCachePrepare(t *testing.T) {
	db := &DBStub{}
	sc := NewStmtCache(db)
	query := "SELECT 1"

	sc.Prepare(query)
	assert.Equal(t, query, db.LastPrepareSql)

	sc.Prepare(query)
	assert.Equal(t, 1, db.PrepareCount, "expected 1 Prepare, got %d", db.PrepareCount)

	// clear statement cache
	assert.Nil(t, sc.Clear())

	// should prepare the query again
	sc.Prepare(query)
	assert.Equal(t, 2, db.PrepareCount, "expected 2 Prepare, got %d", db.PrepareCount)
}
