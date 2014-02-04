package squirrel

import (
	"reflect"
	"testing"
)

func TestWherePartsToSql(t *testing.T) {
	parts := []wherePart{
		newWherePart("x = ?", 1),
		newWherePart(nil),
		newWherePart(Eq{"y": 2}),
	}
	sql, args, _ := wherePartsToSql(parts)
	expect := "x = ? AND y = ?"
	expectArgs := []interface{}{1, 2}
	if sql != expect {
		t.Errorf("expected %#v, got %#v", expect, sql)
	}
	if !reflect.DeepEqual(args, expectArgs) {
		t.Errorf("expected %#v, got %#v", expectArgs, args)
	}
}

func TestWherePartsToSqlErr(t *testing.T) {
	_, _, err := wherePartsToSql([]wherePart{newWherePart(1)})
	if err == nil {
		t.Errorf("expected error, got none")
	}
}

func TestWherePartNil(t *testing.T) {
	sql, _, _ := newWherePart(nil).ToSql()
	expect := ""
	if sql != expect {
		t.Errorf("expected %#v, got %#v", expect, sql)
	}
}

func TestWherePartErr(t *testing.T) {
	_, _, err := newWherePart(1).ToSql()
	if err == nil {
		t.Errorf("expected error, got none")
	}
}

func TestWherePartString(t *testing.T) {
	sql, args, _ := newWherePart("x = ?", 1).ToSql()
	expect := "x = ?"
	expectArgs := []interface{}{1}
	if sql != "x = ?" {
		t.Errorf("expected %#v, got %#v", expect, sql)
	}
	if !reflect.DeepEqual(args, expectArgs) {
		t.Errorf("expected %#v, got %#v", expectArgs, args)
	}
}

func TestWherePartMap(t *testing.T) {
	test := func(pred interface{}) {
		sql, _, _ := newWherePart(pred).ToSql()
		expect := []string{"x = ? AND y = ?", "y = ? AND x = ?"}
		if sql != expect[0] && sql != expect[1] {
			t.Errorf("expected one of %#v, got %#v", expect, sql)
		}
	}
	m := map[string]interface{}{"x": 1, "y": 2}
	test(m)
	test(Eq(m))
}

func TestWherePartMapNil(t *testing.T) {
	sql, _, _ := newWherePart(Eq{"x": nil}).ToSql()
	expect := "x IS NULL"
	if sql != expect {
		t.Errorf("expected %#v, got %#v", expect, sql)
	}
}

func TestWherePartMapSlice(t *testing.T) {
	sql, _, _ := newWherePart(Eq{"x": []int{1, 2}}).ToSql()
	expect := "x IN (?,?)"
	if sql != expect {
		t.Errorf("expected %#v, got %#v", expect, sql)
	}
}
