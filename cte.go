package squirrel

import (
	"bytes"
	"strings"
)

// CTE represents a single common table expression. They are composed of an alias, a few optional components, and a data manipulation statement, though exactly what sort of statement depends on the database system you're using. MySQL, for example, only allows SELECT statements; others, like PostgreSQL, permit INSERTs, UPDATEs, and DELETEs.
// The optional components supported by this fork of Squirrel include:
// * a list of columns
// * the keyword RECURSIVE, the use of which may place additional constraints on the data manipulation statement
type CTE struct {
	Alias      string
	ColumnList []string
	Recursive  bool
	Expression Sqlizer
}

// ToSql builds the SQL for a CTE
func (c CTE) ToSql() (string, []interface{}, error) {

	var buf bytes.Buffer

	if c.Recursive {
		buf.WriteString("RECURSIVE ")
	}

	buf.WriteString(c.Alias)

	if len(c.ColumnList) > 0 {
		buf.WriteString("(")
		buf.WriteString(strings.Join(c.ColumnList, ", "))
		buf.WriteString(")")
	}

	buf.WriteString(" AS (")
	sql, args, err := c.Expression.ToSql()
	if err != nil {
		return "", []interface{}{}, err
	}
	buf.WriteString(sql)
	buf.WriteString(")")

	return buf.String(), args, nil
}
