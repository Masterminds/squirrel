package squirrel

import (
	"reflect"
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
