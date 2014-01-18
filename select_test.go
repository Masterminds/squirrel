package squirrel

import (
	"fmt"
	"testing"
)

func TestSelectBuilderToSql(t *testing.T) {
	sql, args, err := 
		Select("x", "y", "z").
		Distinct().
		From("some_table").
		Where("w = ?", 1).
		Where(Eq{"x": 2, "y": 3}).
		Where(Eq{"z": []int{4,5,6}}).
		GroupBy("foo").
		Having("foo = bar").
		OrderBy("x").
		Limit(1).
		Offset(2).
		ToSql()
	fmt.Println(sql, args, err)
}
