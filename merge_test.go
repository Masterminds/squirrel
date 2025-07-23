package squirrel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeBuilderToSql(t *testing.T) {
	b := Merge("a").
		Prefix("WITH prefix AS ?", 0).
		Values(1, 2).
		ValuesAlias("vals").
		Columns("b", "c").
		On("a.b = vals.b AND a.c = vals.c").
		When("MATCHED THEN UPDATE SET b = vals.b, c = vals.c").
		When("NOT MATCHED THEN INSERT (b, c) VALUES (vals.b, vals.c)").
		Suffix("RETURNING a.b")

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSQL :=
		"WITH prefix AS ? MERGE INTO a USING (VALUES (?,?)) AS vals (b,c) ON a.b = vals.b AND a.c = vals.c " +
			"WHEN MATCHED THEN UPDATE SET b = vals.b, c = vals.c " +
			"WHEN NOT MATCHED THEN INSERT (b, c) VALUES (vals.b, vals.c) RETURNING a.b"
	assert.Equal(t, expectedSQL, sql)

	expectedArgs := []interface{}{0, 1, 2}
	assert.Equal(t, expectedArgs, args)
}

func TestMergeBuilderToSqlErr(t *testing.T) {
	_, _, err := Merge("").Values(1).ToSql()
	assert.Error(t, err)

	_, _, err = Merge("x").ToSql()
	assert.Error(t, err)
}

func TestMergeBuilderMustSql(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestMergeBuilderMustSql should have panicked!")
		}
	}()
	Merge("").MustSql()
}

func TestMergeBuilderPlaceholders(t *testing.T) {
	b := Merge("test").Values(1, 2)

	sql, _, _ := b.PlaceholderFormat(Question).ToSql()
	assert.Equal(t, "MERGE INTO test USING (VALUES (?,?))", sql)

	sql, _, _ = b.PlaceholderFormat(Dollar).ToSql()
	assert.Equal(t, "MERGE INTO test USING (VALUES ($1,$2))", sql)
}

func TestMergeBuilderRunners(t *testing.T) {
	db := &DBStub{}
	b := Merge("test").Values(1).RunWith(db)

	expectedSQL := "MERGE INTO test USING (VALUES (?))"

	b.Exec()
	assert.Equal(t, expectedSQL, db.LastExecSql)
}

func TestMergeBuilderNoRunner(t *testing.T) {
	b := Merge("test").Values(1)

	_, err := b.Exec()
	assert.Equal(t, RunnerNotSet, err)
}

func TestMergeBuilderSetMap(t *testing.T) {
	b := Merge("table").SetMap(Eq{"field1": 1, "field2": 2, "field3": 3})

	sql, args, err := b.ToSql()
	assert.NoError(t, err)

	expectedSQL := "MERGE INTO table USING (VALUES (?,?,?))(field1,field2,field3)"
	assert.Equal(t, expectedSQL, sql)

	expectedArgs := []interface{}{1, 2, 3}
	assert.Equal(t, expectedArgs, args)
}

func TestMergeBuilderSelect(t *testing.T) {
	sb := Select("field1").From("table1").Where(Eq{"field1": 1})
	ib := Merge("table2").ValuesAlias("vals").On("table2.field1 = vals.field1").Columns("field1").Select(sb)

	sql, args, err := ib.ToSql()
	assert.NoError(t, err)

	expectedSQL := "MERGE INTO table2 USING (SELECT field1 FROM table1 WHERE field1 = ?) AS vals (field1) ON table2.field1 = vals.field1"
	assert.Equal(t, expectedSQL, sql)

	expectedArgs := []interface{}{1}
	assert.Equal(t, expectedArgs, args)
}
