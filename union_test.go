package squirrel

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnionPartsAppendToSqlWithString(t *testing.T) {
	parts := []Sqlizer{
		newUnionPart(Select("*").Where("col = ?", 10)),
		newUnionPart("select * from TEST where col = ", "hello"),
	}

	sql := &bytes.Buffer{}
	args, _ := appendToSql(parts, sql, " UNION ", []interface{}{})
	assert.Equal(t, "SELECT * WHERE col = ? UNION select * from TEST where col = ", sql.String())
	assert.Equal(t, []interface{}{10, "hello"}, args)
}

func TestUnionPartsAppendToSqlErr(t *testing.T) {
	parts := []Sqlizer{newUnionPart(1)}
	_, err := appendToSql(parts, &bytes.Buffer{}, "", []interface{}{})
	assert.Error(t, err)
}
