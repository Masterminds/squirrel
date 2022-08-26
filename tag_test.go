package squirrel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDBTAGOK(t *testing.T) {
	type TestStruct struct {
		Name     string `db:"db_name_field"`
		Lastname string
	}
	st := TestStruct{
		Name:     "Testing",
		Lastname: "More Testing",
	}
	expected := "db_name_field"

	result, err := DatabaseFieldName(st, "Name")
	assert.Nil(t, err)
	assert.Equal(t, result, expected)
}

func TestDBTAGNoField(t *testing.T) {
	type TestStruct struct {
		Name     string `db:"db_name_field"`
		Lastname string
	}
	st := TestStruct{
		Name:     "Testing",
		Lastname: "More Testing",
	}
	_, err := DatabaseFieldName(st, "TEST")
	assert.NotNil(t, err)
	assert.Equal(t, err, errFieldNotFound)
}

func TestDBTAGNoTag(t *testing.T) {
	type TestStruct struct {
		Name     string `db:"db_name_field"`
		Lastname string
	}
	st := TestStruct{
		Name:     "Testing",
		Lastname: "More Testing",
	}
	_, err := DatabaseFieldName(st, "Lastname")
	assert.NotNil(t, err)
	assert.Equal(t, err, errEmptyDBTag)
}

func TestDBTAGNoStruct(t *testing.T) {
	testing := "test"

	_, err := DatabaseFieldName(testing, "Field")
	assert.NotNil(t, err)
	assert.Equal(t, err, errNotAStruct)
}

func TestMarshallDBOK(t *testing.T) {

	type TestStruct struct {
		Name        string `db:"db_table.name"`
		Lastname    string
		Age         int   `json:"age" db:"db_table.age"`
		IsSomething *bool `db:"db_table.is_something"`
	}

	type NestedStruct struct {
		Test string `db:"db_table.test"`
		TestStruct
	}

	boolValue := true
	st := NestedStruct{
		Test: "test",
		TestStruct: TestStruct{
			Name:        "Testing",
			Lastname:    "More Testing",
			Age:         15,
			IsSomething: &boolValue,
		},
	}
	expected := map[string]interface{}{}
	expected["db_table.test"] = "test"
	expected["db_table.name"] = "Testing"
	expected["db_table.age"] = 15
	expected["db_table.is_something"] = &boolValue

	result, err := FieldValuesFromInputStruct(st)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestMarshallDBNoStruct(t *testing.T) {
	testing := "test"

	_, err := FieldValuesFromInputStruct(testing)
	assert.NotNil(t, err)
	assert.Equal(t, err, errNotAStruct)
}

func TestMarshallDBEmptyResult(t *testing.T) {
	type TestStruct struct {
		Name     string
		Lastname string
		Age      int `json:"age"`
	}
	st := TestStruct{
		Name:     "Testing",
		Lastname: "More Testing",
		Age:      15,
	}
	expected := map[string]interface{}{}

	result, err := FieldValuesFromInputStruct(st)
	assert.Nil(t, err)
	assert.Equal(t, result, expected)
}
