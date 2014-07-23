package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectBuilderToSql(t *testing.T) {
	b := Select("a", "b").
		Prefix("WITH prefix AS ?", 0).
		Distinct().
		Columns("c").
		Column("IF(d IN ("+Placeholders(3)+"), 1, 0) as stat_column", 1, 2, 3).
		From("e").
		JoinClause("CROSS JOIN j1").
		Join("j2").
		LeftJoin("j3").
		RightJoin("j4").
		Where("f = ?", 4).
		Where(Eq{"g": 5}).
		Where(map[string]interface{}{"h": 6}).
		Where(Eq{"i": []int{7, 8, 9}}).
		GroupBy("j").
		Having("k = l").
		OrderBy("m ASC", "n DESC").
		Limit(10).
		Offset(11).
		Suffix("FETCH FIRST ? ROWS ONLY", 12)

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSql :=
		"WITH prefix AS ? " +
			"SELECT DISTINCT a, b, c, IF(d IN (?,?,?), 1, 0) as stat_column " +
			"FROM e " +
			"CROSS JOIN j1 JOIN j2 LEFT JOIN j3 RIGHT JOIN j4 " +
			"WHERE f = ? AND g = ? AND h = ? AND i IN (?,?,?) " +
			"GROUP BY j HAVING k = l ORDER BY m ASC, n DESC LIMIT 10 OFFSET 11 " +
			"FETCH FIRST ? ROWS ONLY"
	assert.Equal(t, expectedSql, sql)

	expectedArgs := []interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 12}
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
