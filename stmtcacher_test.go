package squirrel

import "testing"

func TestStmtCacherPrepare(t *testing.T) {
	db := &DBStub{}
	sc := NewStmtCacher(db)
	query := "SELECT 1"

	sc.Prepare(query)
	lastSql := db.LastPrepareSql
	if lastSql != query {
		t.Errorf("expected %v, got %v", query, lastSql)
	}

	sc.Prepare(query)
	if db.PrepareCount != 1 {
		t.Errorf("expected 1 Prepare, got %d", db.PrepareCount)
	}
}
