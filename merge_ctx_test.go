//go:build go1.8
// +build go1.8

package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeBuilderContextRunners(t *testing.T) {
	db := &DBStub{}
	b := Merge("test").Values(1).RunWith(db)

	expectedSql := "MERGE INTO test USING (VALUES (?)) ()"

	b.ExecContext(ctx)
	assert.Equal(t, expectedSql, db.LastExecSql)

	b.QueryContext(ctx)
	assert.Equal(t, expectedSql, db.LastQuerySql)

	b.QueryRowContext(ctx)
	assert.Equal(t, expectedSql, db.LastQueryRowSql)

	err := b.ScanContext(ctx)
	assert.NoError(t, err)
}

func TestMergeBuilderContextNoRunner(t *testing.T) {
	b := Merge("test").Values(1)

	_, err := b.ExecContext(ctx)
	assert.Equal(t, RunnerNotSet, err)

	_, err = b.QueryContext(ctx)
	assert.Equal(t, RunnerNotSet, err)

	err = b.ScanContext(ctx)
	assert.Equal(t, RunnerNotSet, err)
}
