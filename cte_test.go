package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalCTE(t *testing.T) {

	cte := CTE{
		Alias:      "cte",
		ColumnList: []string{"abc", "def"},
		Recursive:  false,
		Expression: Select("abc", "def").From("t").Where(Eq{"abc": 1}),
	}

	sql, args, err := cte.ToSql()

	assert.Equal(t, "cte(abc, def) AS (SELECT abc, def FROM t WHERE abc = ?)", sql)
	assert.Equal(t, []interface{}{1}, args)
	assert.Nil(t, err)

}

func TestRecursiveCTE(t *testing.T) {

	// this isn't usually valid SQL, but the point is to test the RECURSIVE part
	cte := CTE{
		Alias:      "cte",
		ColumnList: []string{"abc", "def"},
		Recursive:  true,
		Expression: Select("abc", "def").From("t").Where(Eq{"abc": 1}),
	}

	sql, args, err := cte.ToSql()

	assert.Equal(t, "RECURSIVE cte(abc, def) AS (SELECT abc, def FROM t WHERE abc = ?)", sql)
	assert.Equal(t, []interface{}{1}, args)
	assert.Nil(t, err)

}
