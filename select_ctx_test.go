// +build go1.8

package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectBuilderContextRunners(t *testing.T) {
	db := &DBStub{}
	b := Select("test").RunWith(db)

	expectedSql := "SELECT test"

	b.ExecContext(ctx)
	assert.Equal(t, expectedSql, db.LastExecSql)

	b.QueryContext(ctx)
	assert.Equal(t, expectedSql, db.LastQuerySql)

	b.QueryRowContext(ctx)
	assert.Equal(t, expectedSql, db.LastQueryRowSql)

	err := b.ScanContext(ctx)
	assert.NoError(t, err)
}

func TestSelectBuilderContextNoRunner(t *testing.T) {
	b := Select("test")

	_, err := b.ExecContext(ctx)
	assert.Equal(t, RunnerNotSet, err)

	_, err = b.QueryContext(ctx)
	assert.Equal(t, RunnerNotSet, err)

	err = b.ScanContext(ctx)
	assert.Equal(t, RunnerNotSet, err)
}
