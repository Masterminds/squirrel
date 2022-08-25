package squirrel

import (
	"errors"
	"reflect"
)

var errFieldNotFound = errors.New("field could not be found in struct")
var errEmptyDBTag = errors.New("tag db is empty")
var errNotAStruct = errors.New("not a struct")

// DBTAG Gets the value of the tag "db" in the struct passed field name
// returns an error if the field does not exist
func DBTAG(st any, field string) (string, error) {
	if reflect.TypeOf(st).Kind() != reflect.Struct {
		return "", errNotAStruct
	}
	reflection := reflect.TypeOf(st)
	stField, ok := reflection.FieldByName(field)
	if !ok {
		return "", errFieldNotFound
	}

	dbFieldName := stField.Tag.Get("db")
	if dbFieldName == "" {
		return "", errEmptyDBTag
	}

	return dbFieldName, nil
}

// MarshallDB returns a map with the Database Field Name & its Value
// from the struct passed as parameter.
// if no 'db' tag is found in a field, that field wont be in the retuned map
func MarshallDB(st any) (map[string]interface{}, error) {
	values := map[string]interface{}{}
	if reflect.TypeOf(st).Kind() != reflect.Struct {
		return values, errNotAStruct
	}
	reflection := reflect.ValueOf(st)
	reflectionType := reflection.Type()
	for i := 0; i < reflection.NumField(); i++ {
		if reflection.Field(i).Kind() == reflect.Struct {
			// support nested struct
			nestedFields, err := MarshallDB(reflection.Field(i).Interface())
			if err != nil {
				return values, err
			}
			for k, v := range nestedFields {
				values[k] = v
			}
		} else {
			dbfield, err := DBTAG(st, reflectionType.Field(i).Name)
			if err != nil {
				if err == errEmptyDBTag {
					continue
				}
				return values, err
			}
			values[dbfield] = reflection.Field(i).Interface()
		}
	}

	return values, nil
}
