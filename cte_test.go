package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithAsQuery_OneSubquery(t *testing.T) {
	w := With("lab").As(
		Select("col").From("tab").
			Where("simple").
			Where("NOT hard"),
	).Select(
		Select("col").
			From("lab"),
	)
	q, _, err := w.ToSql()
	assert.NoError(t, err)

	expectedSql := "WITH lab AS (\n" +
		"SELECT col FROM tab WHERE simple AND NOT hard\n" +
		")\n" +
		"SELECT col FROM lab"
	assert.Equal(t, expectedSql, q)

	w = WithRecursive("lab").As(
		Select("col").From("tab").
			Where("simple").
			Where("NOT hard"),
	).Select(Select("col").
		From("lab"),
	)
	q, _, err = w.ToSql()
	assert.NoError(t, err)

	expectedSql = "WITH RECURSIVE lab AS (\n" +
		"SELECT col FROM tab WHERE simple AND NOT hard\n" +
		")\n" +
		"SELECT col FROM lab"
	assert.Equal(t, expectedSql, q)
}

func TestWithAsQuery_TwoSubqueries(t *testing.T) {
	w := With("lab_1").As(
		Select("col_1", "col_common").From("tab_1").
			Where("simple").
			Where("NOT hard"),
	).Cte("lab_2").As(
		Select("col_2", "col_common").From("tab_2"),
	).Select(Select("col_1", "col_2", "col_common").
		From("lab_1").Join("lab_2 ON lab_1.col_common = lab_2.col_common"),
	)
	q, _, err := w.ToSql()
	assert.NoError(t, err)

	expectedSql := "WITH lab_1 AS (\n" +
		"SELECT col_1, col_common FROM tab_1 WHERE simple AND NOT hard\n" +
		"), lab_2 AS (\n" +
		"SELECT col_2, col_common FROM tab_2\n" +
		")\n" +
		"SELECT col_1, col_2, col_common FROM lab_1 JOIN lab_2 ON lab_1.col_common = lab_2.col_common"
	assert.Equal(t, expectedSql, q)
}

func TestWithAsQuery_ManySubqueries(t *testing.T) {
	w := With("lab_1").As(
		Select("col_1", "col_common").From("tab_1").
			Where("simple").
			Where("NOT hard"),
	).Cte("lab_2").As(
		Select("col_2", "col_common").From("tab_2"),
	).Cte("lab_3").As(
		Select("col_3", "col_common").From("tab_3"),
	).Cte("lab_4").As(
		Select("col_4", "col_common").From("tab_4"),
	).Select(
		Select("col_1", "col_2", "col_3", "col_4", "col_common").
			From("lab_1").Join("lab_2 ON lab_1.col_common = lab_2.col_common").
			Join("lab_3 ON lab_1.col_common = lab_3.col_common").
			Join("lab_4 ON lab_1.col_common = lab_4.col_common"),
	)
	q, _, err := w.ToSql()
	assert.NoError(t, err)

	expectedSql := "WITH lab_1 AS (\n" +
		"SELECT col_1, col_common FROM tab_1 WHERE simple AND NOT hard\n" +
		"), lab_2 AS (\n" +
		"SELECT col_2, col_common FROM tab_2\n" +
		"), lab_3 AS (\n" +
		"SELECT col_3, col_common FROM tab_3\n" +
		"), lab_4 AS (\n" +
		"SELECT col_4, col_common FROM tab_4\n" +
		")\n" +
		"SELECT col_1, col_2, col_3, col_4, col_common FROM lab_1 JOIN lab_2 ON lab_1.col_common = lab_2.col_common JOIN lab_3 ON lab_1.col_common = lab_3.col_common JOIN lab_4 ON lab_1.col_common = lab_4.col_common"
	assert.Equal(t, expectedSql, q)
}

func TestWithAsQuery_Insert(t *testing.T) {
	w := With("lab").As(
		Select("col").From("tab").
			Where("simple").
			Where("NOT hard"),
	).Insert(Insert("ins_tab").Columns("ins_col").Select(Select("col").From("lab")))
	q, _, err := w.ToSql()
	assert.NoError(t, err)

	expectedSql := "WITH lab AS (\n" +
		"SELECT col FROM tab WHERE simple AND NOT hard\n" +
		")\n" +
		"INSERT INTO ins_tab (ins_col) SELECT col FROM lab"
	assert.Equal(t, expectedSql, q)
}

func TestWithAsQuery_Update(t *testing.T) {
	w := With("lab").As(
		Select("col", "common_col").From("tab").
			Where("simple").
			Where("NOT hard"),
	).Update(
		Update("upd_tab, lab").
			Set("upd_col", Expr("lab.col")).
			Where("common_col = lab.common_col"),
	)

	q, _, err := w.ToSql()
	assert.NoError(t, err)

	expectedSql := "WITH lab AS (\n" +
		"SELECT col, common_col FROM tab WHERE simple AND NOT hard\n" +
		")\n" +
		"UPDATE upd_tab, lab SET upd_col = lab.col WHERE common_col = lab.common_col"

	assert.Equal(t, expectedSql, q)
}
