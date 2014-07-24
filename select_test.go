package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectBuilderToSql(t *testing.T) {
	//see https://code.google.com/p/go-wiki/wiki/InterfaceSlice
	var dataSlice []int = []int{4, 5, 6}
	var interfaceSlice []interface{} = make([]interface{}, len(dataSlice))
	for i, d := range dataSlice {
		interfaceSlice[i] = d
	}

	b := Select("a", "b").
		Prefix("WITH prefix AS ?", 0).
		Distinct().
		Columns("c1").
    	Columns("c2", "c3", "c4").
        Column("IF(c5 IN ("+Placeholders(3)+"), 1, 0) as stat_column", 1, 2, 3).
        Column("IF(c6 IN ("+Placeholders(3)+"), 1, 0) as stat_column2", interfaceSlice...).
        From("d").
		JoinClause("CROSS JOIN j1").
		Join("j2").
		LeftJoin("j3").
		RightJoin("j4").
		Where("e = ?", 1).
		Where(Eq{"f": 2}).
		Where(map[string]interface{}{"g": 3}).
		Where(Eq{"h": []int{4, 5, 6}}).
		GroupBy("i").
		Having("j = k").
		OrderBy("l ASC", "m DESC").
		Limit(7).
		Offset(8).
		Suffix("FETCH FIRST ? ROWS ONLY", 7)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql :=
		"WITH prefix AS ? " +
			"SELECT DISTINCT a, b, c1, с2, с3, IF(с4 IN (?,?,?), 1, 0) as stat_column, IF(с5 IN (?,?,?), 1, 0) as stat_column2 " + 
            "FROM d " +
			"CROSS JOIN j1 JOIN j2 LEFT JOIN j3 RIGHT JOIN j4 " +
			"WHERE e = ? AND f = ? AND g = ? AND h IN (?,?,?) " +
			"GROUP BY i HAVING j = k ORDER BY l ASC, m DESC LIMIT 7 OFFSET 8 " +
			"FETCH FIRST ? ROWS ONLY"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 15}
	assert.Equal(t, expectedArgs, args)
}

func TestSelectBuilderToSqlErr(t *testing.T) {
	_, _, err := Select().From("x").ToSql()
	assert.Error(t, err)
}

func TestSelectBuilderPlaceholders(t *testing.T) {
	b := Select("test").Where("x = ? AND y = ?")

	sql, _, _ := b.PlaceholderFormat(Question).ToSql()
	assert.Equal(t, "SELECT test WHERE x = ? AND y = ?", sql)

	sql, _, _ = b.PlaceholderFormat(Dollar).ToSql()
	assert.Equal(t, "SELECT test WHERE x = $1 AND y = $2", sql)
}

func TestSelectBuilderRunners(t *testing.T) {
	db := &DBStub{}
	b := Select("test").RunWith(db)

	expectedSql := "SELECT test"

	b.Exec()
	assert.Equal(t, expectedSql, db.LastExecSql)

	b.Query()
	assert.Equal(t, expectedSql, db.LastQuerySql)

	b.QueryRow()
	assert.Equal(t, expectedSql, db.LastQueryRowSql)

	err := b.Scan()
	assert.NoError(t, err)
}

func TestSelectBuilderNoRunner(t *testing.T) {
	b := Select("test")

	_, err := b.Exec()
	assert.Equal(t, RunnerNotSet, err)

	_, err = b.Query()
	assert.Equal(t, RunnerNotSet, err)

	err = b.Scan()
	assert.Equal(t, RunnerNotSet, err)
}
