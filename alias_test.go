package squirrel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAliasNestedSelectPlaceholders(t *testing.T) {
	tests := []struct {
		name   string
		format PlaceholderFormat
		bind   [4]string
	}{
		{"Question", Question, [4]string{"?", "?", "?", "?"}},
		{"Dollar", Dollar, [4]string{"$1", "$2", "$3", "$4"}},
		{"Colon", Colon, [4]string{":1", ":2", ":3", ":4"}},
		{"AtP", AtP, [4]string{"@p1", "@p2", "@p3", "@p4"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			builder := StatementBuilder.PlaceholderFormat(test.format)
			first := Alias(builder.Select("COUNT(*)").From("items").Where(Eq{"state": "first"}), "first_total")
			second := Alias(builder.Select("COUNT(*)").From("items").Where(Eq{"state": "second"}), "second_total")

			sql, args, err := builder.Select().
				Column("?", "before").
				Column(first).
				Column(second).
				From("items").
				Where(Eq{"owner": "outer"}).
				ToSql()
			assert.NoError(t, err)
			want := "SELECT " + test.bind[0] +
				", (SELECT COUNT(*) FROM items WHERE state = " + test.bind[1] + ") AS first_total" +
				", (SELECT COUNT(*) FROM items WHERE state = " + test.bind[2] + ") AS second_total" +
				" FROM items WHERE owner = " + test.bind[3]
			assert.Equal(t, want, sql)
			assert.Equal(t, []interface{}{"before", "first", "second", "outer"}, args)

			// A standalone alias still uses its child's placeholder format.
			sql, args, err = first.ToSql()
			assert.NoError(t, err)
			assert.Equal(t, "(SELECT COUNT(*) FROM items WHERE state = "+test.bind[0]+") AS first_total", sql)
			assert.Equal(t, []interface{}{"first"}, args)
		})
	}
}

func TestAliasNestedSqlizer(t *testing.T) {
	value := Alias(aliasTestSqlizer{sql: "COALESCE(?, ?)", args: []interface{}{"first", "second"}}, "value")
	sql, args, err := Select().Column(value).From("items").Where(Eq{"id": 3}).PlaceholderFormat(Dollar).ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT (COALESCE($1, $2)) AS value FROM items WHERE id = $3", sql)
	assert.Equal(t, []interface{}{"first", "second", 3}, args)
}

func TestAliasSqlizerError(t *testing.T) {
	wantErr := errors.New("sqlizer failed")
	wantArgs := []interface{}{"argument"}
	value := Alias(aliasTestSqlizer{sql: "partial", args: wantArgs, err: wantErr}, "value")
	for _, build := range []func() (string, []interface{}, error){
		value.ToSql,
		func() (string, []interface{}, error) { return nestedToSql(value) },
	} {
		sql, args, err := build()
		assert.Equal(t, "partial", sql)
		assert.Equal(t, wantArgs, args)
		assert.Equal(t, wantErr, err)
	}
	_, _, err := Select().Column(value).ToSql()
	assert.Equal(t, wantErr, err)
}

type aliasTestSqlizer struct {
	sql  string
	args []interface{}
	err  error
}

func (s aliasTestSqlizer) ToSql() (string, []interface{}, error) {
	return s.sql, s.args, s.err
}
