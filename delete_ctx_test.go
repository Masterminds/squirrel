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
}

func TestDeleteBuilderContextNoRunner(t *testing.T) {
	b := Delete("test")

	_, err := b.ExecContext(ctx)
	assert.Equal(t, RunnerNotSet, err)
}
