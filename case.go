package squirrel

import (
	"bytes"
	"errors"

	"github.com/lann/builder"
)

func init() {
	builder.Register(CaseBuilder{}, caseData{})
}

// sqlizerBuffer is a helper that allows to write many Sqlizers one by one
// without constant checks for errors that may come from Sqlizer
type sqlizerBuffer struct {
	bytes.Buffer
	args []interface{}
	err  error
}

// WriteSql converts Sqlizer to SQL strings and writes it to buffer
func (b *sqlizerBuffer) WriteSql(item Sqlizer, serializer Serializer) {
	if b.err != nil {
		return
	}

	var str string
	var args []interface{}
	str, args, b.err = item.ToSqlWithSerializer(serializer)

	if b.err != nil {
		return
	}

	b.WriteString(str)
	b.WriteByte(' ')
	b.args = append(b.args, args...)
}

func (b *sqlizerBuffer) ToSql() (string, []interface{}, error) {
	return b.String(), b.args, b.err
}

// whenPart is a helper structure to describe SQLs "WHEN ... THEN ..." expression
type whenPart struct {
	when Sqlizer
	then Sqlizer
}

func newWhenPart(when interface{}, then interface{}) whenPart {
	return whenPart{newPart(when), newPart(then)}
}

// caseData holds all the data required to build a CASE SQL construct
type caseData struct {
	What      Sqlizer
	WhenParts []whenPart
	Else      Sqlizer
}

// ToSql implements Sqlizer
func (d *caseData) ToSql() (sqlStr string, args []interface{}, err error) {
	return d.ToSqlWithSerializer(DefaultSerializer{})
}

// ToSql implements Sqlizer
func (d *caseData) ToSqlWithSerializer(serializer Serializer) (sqlStr string, args []interface{}, err error) {
	if len(d.WhenParts) == 0 {
		err = errors.New("case expression must contain at lease one WHEN clause")

		return
	}

	return serializer.Case(*d)
}

// CaseBuilder builds SQL CASE construct which could be used as parts of queries.
type CaseBuilder builder.Builder

// ToSql builds the query into a SQL string and bound args with the default serializer.
func (b CaseBuilder) ToSql() (sqlStr string, args []interface{}, err error) {
	return b.ToSqlWithSerializer(DefaultSerializer{})
}

// ToSql builds the query into a SQL string and bound args with a specific serializer.
func (b CaseBuilder) ToSqlWithSerializer(serializer Serializer) (string, []interface{}, error) {
	data := builder.GetStruct(b).(caseData)
	return data.ToSqlWithSerializer(serializer)
}

// what sets optional value for CASE construct "CASE [value] ..."
func (b CaseBuilder) what(expr interface{}) CaseBuilder {
	return builder.Set(b, "What", newPart(expr)).(CaseBuilder)
}

// When adds "WHEN ... THEN ..." part to CASE construct
func (b CaseBuilder) When(when interface{}, then interface{}) CaseBuilder {
	// TODO: performance hint: replace slice of WhenPart with just slice of parts
	// where even indices of the slice belong to "when"s and odd indices belong to "then"s
	return builder.Append(b, "WhenParts", newWhenPart(when, then)).(CaseBuilder)
}

// What sets optional "ELSE ..." part for CASE construct
func (b CaseBuilder) Else(expr interface{}) CaseBuilder {
	return builder.Set(b, "Else", newPart(expr)).(CaseBuilder)
}
