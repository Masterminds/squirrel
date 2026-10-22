package squirrel

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilValuerToSql(t *testing.T) {
	values := []struct {
		name  string
		value interface{}
	}{
		{"NullString", (*sql.NullString)(nil)},
		{"NullInt64", (*sql.NullInt64)(nil)},
	}
	for _, value := range values {
		for _, op := range valuerOperators() {
			t.Run(value.name+"/"+op.name, func(t *testing.T) {
				assert.NotPanics(t, func() {
					sql, args, err := op.build(value.value).ToSql()
					assert.Empty(t, args)
					if op.nullError != "" {
						if assert.Error(t, err) {
							assert.Equal(t, op.nullError, err.Error())
						}
					} else {
						assert.NoError(t, err)
						assert.Equal(t, "name "+op.nullOperator+" NULL", sql)
					}
				})
			})
		}
	}
}

func TestValuerToSql(t *testing.T) {
	value := sql.NullString{String: "Name", Valid: true}
	values := []struct {
		name  string
		value interface{}
	}{
		{"value", value},
		{"pointer", &value},
		{"nilPointerReceiver", (*nilPointerValuer)(nil)},
		{"pointerReceiver", &nilPointerValuer{}},
	}
	for _, value := range values {
		for _, op := range valuerOperators() {
			t.Run(value.name+"/"+op.name, func(t *testing.T) {
				sql, args, err := op.build(value.value).ToSql()
				assert.NoError(t, err)
				assert.Equal(t, "name "+op.name+" ?", sql)
				assert.Equal(t, []interface{}{"Name"}, args)
			})
		}
	}
}

func TestValuerErrorToSql(t *testing.T) {
	for _, op := range valuerOperators() {
		t.Run(op.name, func(t *testing.T) {
			_, _, err := op.build((*nilPointerErrorValuer)(nil)).ToSql()
			assert.Equal(t, errNilPointerValuer, err)
		})
	}
}

type nilPointerValuer struct{}

func (*nilPointerValuer) Value() (driver.Value, error) {
	return "Name", nil
}

var errNilPointerValuer = errors.New("nil pointer Valuer error")

type nilPointerErrorValuer struct{}

func (*nilPointerErrorValuer) Value() (driver.Value, error) {
	return nil, errNilPointerValuer
}

func valuerOperators() []struct {
	name         string
	build        func(interface{}) Sqlizer
	nullOperator string
	nullError    string
} {
	return []struct {
		name         string
		build        func(interface{}) Sqlizer
		nullOperator string
		nullError    string
	}{
		{"=", func(v interface{}) Sqlizer { return Eq{"name": v} }, "IS", ""},
		{"<>", func(v interface{}) Sqlizer { return NotEq{"name": v} }, "IS NOT", ""},
		{"LIKE", func(v interface{}) Sqlizer { return Like{"name": v} }, "", "cannot use null with like operators"},
		{"NOT LIKE", func(v interface{}) Sqlizer { return NotLike{"name": v} }, "", "cannot use null with like operators"},
		{"ILIKE", func(v interface{}) Sqlizer { return ILike{"name": v} }, "", "cannot use null with like operators"},
		{"NOT ILIKE", func(v interface{}) Sqlizer { return NotILike{"name": v} }, "", "cannot use null with like operators"},
		{"<", func(v interface{}) Sqlizer { return Lt{"name": v} }, "", "cannot use null with less than or greater than operators"},
		{"<=", func(v interface{}) Sqlizer { return LtOrEq{"name": v} }, "", "cannot use null with less than or greater than operators"},
		{">", func(v interface{}) Sqlizer { return Gt{"name": v} }, "", "cannot use null with less than or greater than operators"},
		{">=", func(v interface{}) Sqlizer { return GtOrEq{"name": v} }, "", "cannot use null with less than or greater than operators"},
	}
}
