// +build go1.8

package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteBuilderContextRunners(t *testing.T) {
	db := &DBStub{}
	b := Delete("test").Where("x = ?", 1).RunWith(db)

	expectedSql := "DELETE FROM test WHERE x = ?"

	b.ExecContext(ctx)
	assert.Equal(t, expectedSql, db.LastExecSql)

	b.QueryContext(ctx)
	assert.Equal(t, expectedSql, db.LastQuerySql)

	b.QueryRowContext(ctx)
	assert.Equal(t, expectedSql, db.LastQueryRowSql)

	err := b.ScanContext(ctx)
	assert.NoError(t, err)
}

func TestDeleteBuilderContextNoRunner(t *testing.T) {
	b := Delete("test").Where("x != ?", 0).Suffix("RETURNING x")

	_, err := b.ExecContext(ctx)
	assert.Equal(t, RunnerNotSet, err)

	_, err = b.QueryContext(ctx)
	assert.Equal(t, RunnerNotSet, err)

	err = b.ScanContext(ctx)
	assert.Equal(t, RunnerNotSet, err)
}
