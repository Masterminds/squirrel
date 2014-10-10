package squirrel

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/lann/builder"
)

type createData struct {
	PlaceholderFormat PlaceholderFormat
	Table             string
	Columns           []string
}

// Builder

// SelectBuilder builds SQL SELECT statements.
type CreateBuilder builder.Builder

func init() {
	builder.Register(CreateBuilder{}, createData{})
}

func (d createData) ToSql() (sqlStr string, args []interface{}, err error) {
	if len(d.Columns) == 0 {
		err = fmt.Errorf("create statements must have at least column")
		return
	}

	sql := &bytes.Buffer{}

	sql.WriteString("CREATE TABLE ")
	sql.WriteString(d.Table)

	sql.WriteString(" (")
	sql.WriteString(strings.Join(d.Columns, ","))
	sql.WriteString(")")

	return sql.String(), nil, nil
}

// SQL methods

// ToSql builds the query into a SQL string and bound args.
func (b CreateBuilder) ToSql() (string, []interface{}, error) {
	data := builder.GetStruct(b).(createData)
	return data.ToSql()
}

// Table adds table name to query.
func (b CreateBuilder) Table(table string) CreateBuilder {
	return builder.Set(b, "Table", table).(CreateBuilder)
}

func (b CreateBuilder) Column(name, definition string, options ...string) CreateBuilder {
	column := fmt.Sprintf("%s %s", name, definition)
	if len(options) > 0 {
		column = fmt.Sprintf("%s %s", column, strings.Join(options, " "))
	}
	return builder.Append(b, "Columns", column).(CreateBuilder)
}
