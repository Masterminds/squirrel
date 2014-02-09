package squirrel

import "testing"

func TestQuestion(t *testing.T) {
	sql := "x = ? AND y = ?"
	s, _ := Question.ReplacePlaceholders(sql)
	if s != sql {
		t.Errorf("expected %v, got %v", sql, s)
	}
}

func TestDollar(t *testing.T) {
	sql := "x = ? AND y = ?"
	s, _ := Dollar.ReplacePlaceholders(sql)
	expect := "x = $1 AND y = $2"
	if s != expect {
		t.Errorf("expected %v, got %v", expect, s)
	}
}
