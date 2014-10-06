package squirrel

import (
	"bytes"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestFromPartsAppendToSqlWithString(t *testing.T) {
	parts := []Sqlizer{
		newFromPart("TABLE_NAME"),
	}

	sql := &bytes.Buffer{}
	args, _ := appendToSql(parts, sql, "", []interface{}{})
	assert.Equal(t, "TABLE_NAME", sql.String())
	assert.Equal(t, []interface{}{}, args)
}

func TestFromPartAppendToSqlWithSelect(t *testing.T) {
	parts := []Sqlizer{
		newFromPart(SubQuerySelect("*").Where(Eq{"col":"value"})),
	}

	sql := &bytes.Buffer{}
	args, _ := appendToSql(parts, sql, "", []interface{}{})
	assert.Equal(t, "( SELECT * WHERE col = ? )", sql.String())
	assert.Equal(t, []interface{}{"value"}, args)
}

func TestFromPartAppendToSqlErr(t *testing.T) {
	parts := []Sqlizer{
		newFromPart(1),
	}

	_, err := appendToSql(parts, &bytes.Buffer{}, "", []interface{}{})
	assert.Error(t, err)
}

func TestFromPartToSqlErr(t *testing.T) {
	_, _, err := newFromPart(1).ToSql()
	assert.Error(t, err)
}
