// +build go1.8

package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStmtCacherPrepareContext(t *testing.T) {
	db := &DBStub{}
	sc := NewStmtCacher(db)
	query := "SELECT 1"

	sc.PrepareContext(ctx, query)
	assert.Equal(t, query, db.LastPrepareSql)

	sc.PrepareContext(ctx, query)
	assert.Equal(t, 1, db.PrepareCount, "expected 1 Prepare, got %d", db.PrepareCount)
}
