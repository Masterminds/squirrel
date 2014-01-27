package squirrel

import "testing"

func TestSelect(t *testing.T) {
	sql, _, _ := Select("test").ToSql()
	expectedSql := "SELECT test"
	if sql != sqlStr {
		t.Errorf("expected %v, got %v", expectedSql, sql)
	}
}

func TestStatementBuilder(t *testing.T) {
	db := &DBStub{}
	sb := StatementBuilder.RunWith(db)

	expectedSql := "SELECT test"

	sb.Select("test").Exec()
	sql := db.LastExecSql
	if sql != sqlStr {
		t.Errorf("expected %v, got %v", expectedSql, sql)
	}
}
