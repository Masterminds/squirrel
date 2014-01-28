package squirrel

import (
	"reflect"
	"sort"
	"testing"
)

func TestNewWherePartsNil(t *testing.T) {
	parts := newWhereParts(nil)
	if parts != nil {
		t.Errorf("expected nil, got %v", parts)
	}

	parts = newWhereParts("")
	if parts != nil {
		t.Errorf("expected nil, got %v", parts)
	}
}

type bySql []wherePart
func (a bySql) Len() int           { return len(a) }
func (a bySql) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a bySql) Less(i, j int) bool { return a[i].sql < a[j].sql }

func TestNewWherePartsEqMap(t *testing.T) {
	m := map[string]interface{}{"a": 1, "b": nil, "c": []int{2, 3}}
	eq := Eq(m)
	expected := []wherePart{
		{sql: "a = ?", args: []interface{}{1}},
		{sql: "b IS NULL"},
		{sql: "c IN (?,?)", args: []interface{}{2, 3}},
	}

	check := func(pred interface{}) {
		parts := newWhereParts(pred)
		sort.Sort(bySql(parts))
		if !reflect.DeepEqual(parts, expected) {
			t.Errorf("expected %v, got %v", expected, parts)
		}
	}
	check(m)
	check(eq)
}

func TestNewWherePartsPanic(t *testing.T) {
	var panicVal error
	func() {
		defer func() { panicVal = recover().(error) }()
		newWhereParts(false)
	}()
	if panicVal == nil {
		t.Errorf("expected panic, didn't")
	}
}
