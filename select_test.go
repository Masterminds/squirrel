package squirrel

import (
	"reflect"
	"testing"
)

func TestSelectBuilderToSql(t *testing.T) {
	b := Select("a", "b").
		Distinct().
		Columns("c").
		From("d").
		Where("e = ?", 1).
		Where(Eq{"f": 2, "g": 3}).
		Where(Eq{"h": []int{4,5,6}}).
		GroupBy("i").
		Having("j = k").
		OrderBy("l").
		Limit(7).
		Offset(8)

	sql, args, err := b.ToSql()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedSql :=
		"SELECT DISTINCT a, b, c FROM d " +
		"WHERE e = ? AND f = ? AND g = ? AND h IN (?,?,?) " +
		"GROUP BY i HAVING j = k ORDER BY l LIMIT 7 OFFSET 8"
	if sql != expectedSql {
		t.Errorf("expected %v, got %v", expectedSql, sql)
	}

	expectedArgs := []interface{}{1, 2, 3, 4, 5, 6}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("expected %v, got %v", expectedArgs, args)
	}
}

func TestSelectBuilderToSqlErr(t *testing.T) {
	_, _, err := Select().From("x").ToSql()
	if err == nil {
		t.Error("expected error, got none")
	}
}

func TestSelectBuilderRunners(t *testing.T) {
	db := &DBStub{}
	b := selectWith(db, "test")

	expectedSql := "SELECT test"

	b.Exec()
	sql := db.LastExecSql
	if sql != sqlStr {
		t.Errorf("expected %v, got %v", expectedSql, sql)
	}

	b.Query()
	sql = db.LastQuerySql
	if sql != sqlStr {
		t.Errorf("expected %v, got %v", expectedSql, sql)
	}

	b.QueryRow()
	sql = db.LastQueryRowSql
	if sql != sqlStr {
		t.Errorf("expected %v, got %v", expectedSql, sql)
	}

	err := b.Scan()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestSelectBuilderNoRunner(t *testing.T) {
	b := Select("test")

	_, err := b.Exec()
	if err != RunnerNotSet {
		t.Errorf("expected error %v, got %v", RunnerNotSet, err)
	}

	_, err = b.Query()
	if err != RunnerNotSet {
		t.Errorf("expected error %v, got %v", RunnerNotSet, err)
	}

	err = b.Scan()
	if err != RunnerNotSet {
		t.Errorf("expected error %v, got %v", RunnerNotSet, err)
	}
}
