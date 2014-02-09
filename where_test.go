package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWherePartsToSql(t *testing.T) {
	parts := []wherePart{
		newWherePart("x = ?", 1),
		newWherePart(nil),
		newWherePart(Eq{"y": 2}),
	}
	sql, args, _ := wherePartsToSql(parts)
	assert.Equal(t, "x = ? AND y = ?", sql)
	assert.Equal(t, []interface{}{1, 2}, args)
}

func TestWherePartsToSqlErr(t *testing.T) {
	_, _, err := wherePartsToSql([]wherePart{newWherePart(1)})
	assert.Error(t, err)
}

func TestWherePartNil(t *testing.T) {
	sql, _, _ := newWherePart(nil).ToSql()
	assert.Equal(t, "", sql)
}

func TestWherePartErr(t *testing.T) {
	_, _, err := newWherePart(1).ToSql()
	assert.Error(t, err)
}

func TestWherePartString(t *testing.T) {
	sql, args, _ := newWherePart("x = ?", 1).ToSql()
	assert.Equal(t, "x = ?", sql)
	assert.Equal(t, []interface{}{1}, args)
}

func TestWherePartMap(t *testing.T) {
	test := func(pred interface{}) {
		sql, _, _ := newWherePart(pred).ToSql()
		expect := []string{"x = ? AND y = ?", "y = ? AND x = ?"}
		if sql != expect[0] && sql != expect[1] {
			t.Errorf("expected one of %#v, got %#v", expect, sql)
		}
	}
	m := map[string]interface{}{"x": 1, "y": 2}
	test(m)
	test(Eq(m))
}

func TestWherePartMapNil(t *testing.T) {
	sql, _, _ := newWherePart(Eq{"x": nil}).ToSql()
	assert.Equal(t, "x IS NULL", sql)
}

func TestWherePartMapSlice(t *testing.T) {
	sql, _, _ := newWherePart(Eq{"x": []int{1, 2}}).ToSql()
	assert.Equal(t, "x IN (?,?)", sql)
}
