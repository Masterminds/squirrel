package squirrel

import "testing"

func TestBoundSelect(t *testing.T) {
	db := &DBStub{}
	bound := NewBoundBuilder(db)
	bound.Select("test").Exec()

	if db.LastExecSql != "SELECT test" {
		t.Error("expected db.Exec to be called, but it wasn't")
	}
}
