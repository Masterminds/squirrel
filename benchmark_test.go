// from https://github.com/elgris/golang-sql-builder-benchmark/blob/master/squirrel_benchmark_test.go

package squirrel_test

import (
	"testing"

	"github.com/Masterminds/squirrel"
)

func BenchmarkSquirrelSelectSimple(b *testing.B) {
	for n := 0; n < b.N; n++ {
		squirrel.Select("id").
			From("tickets").
			Where("subdomain_id = ? and (state = ? or state = ?)", 1, "open", "spam").
			ToSql()
	}
}

func BenchmarkSquirrelSelectConditional(b *testing.B) {
	for n := 0; n < b.N; n++ {
		qb := squirrel.Select("id").
			From("tickets").
			Where("subdomain_id = ? and (state = ? or state = ?)", 1, "open", "spam")

		if n%2 == 0 {
			qb = qb.GroupBy("subdomain_id").
				Having("number = ?", 1).
				OrderBy("state").
				Limit(7).
				Offset(8)
		}

		qb.ToSql()
	}
}

func BenchmarkSquirrelSelectComplex(b *testing.B) {
	for n := 0; n < b.N; n++ {
		squirrel.Select("a", "b", "z", "y", "x").
			Distinct().
			From("c").
			Where("d = ? OR e = ?", 1, "wat").
			Where(squirrel.Eq{"f": 2, "x": "hi"}).
			Where(map[string]interface{}{"g": 3}).
			Where(squirrel.Eq{"h": []int{1, 2, 3}}).
			GroupBy("i").
			GroupBy("ii").
			GroupBy("iii").
			Having("j = k").
			Having("jj = ?", 1).
			Having("jjj = ?", 2).
			OrderBy("l").
			OrderBy("l").
			OrderBy("l").
			Limit(7).
			Offset(8).
			ToSql()
	}
}

func BenchmarkSquirrelSelectSubquery(b *testing.B) {
	for n := 0; n < b.N; n++ {
		subSelect := squirrel.Select("id").
			From("tickets").
			Where("subdomain_id = ? and (state = ? or state = ?)", 1, "open", "spam")

		squirrel.Select("a", "b").
			From("c").
			Distinct().
			Column(squirrel.Alias(subSelect, "subq")).
			Where(squirrel.Eq{"f": 2, "x": "hi"}).
			Where(map[string]interface{}{"g": 3}).
			OrderBy("l").
			OrderBy("l").
			Limit(7).
			Offset(8).
			ToSql()

	}
}

func BenchmarkSquirrelSelectMoreComplex(b *testing.B) {
	for i := 0; i < b.N; i++ {

		squirrel.Select("a", "b").
			Prefix("WITH prefix AS ?", 0).
			Distinct().
			Columns("c").
			Column("IF(d IN ("+squirrel.Placeholders(3)+"), 1, 0) as stat_column", 1, 2, 3).
			Column(squirrel.Expr("a > ?", 100)).
			Column(squirrel.Eq{"b": []int{101, 102, 103}}).
			From("e").
			JoinClause("CROSS JOIN j1").
			Join("j2").
			LeftJoin("j3").
			RightJoin("j4").
			Where("f = ?", 4).
			Where(squirrel.Eq{"g": 5}).
			Where(map[string]interface{}{"h": 6}).
			Where(squirrel.Eq{"i": []int{7, 8, 9}}).
			Where(squirrel.Or{squirrel.Expr("j = ?", 10), squirrel.And{squirrel.Eq{"k": 11}, squirrel.Expr("true")}}).
			GroupBy("l").
			Having("m = n").
			OrderBy("o ASC", "p DESC").
			Limit(12).
			Offset(13).
			Suffix("FETCH FIRST ? ROWS ONLY", 14).
			ToSql()
	}
}

//
// Insert benchmark
//
func BenchmarkSquirrelInsert(b *testing.B) {
	for n := 0; n < b.N; n++ {
		squirrel.Insert("mytable").
			Columns("id", "a", "b", "price", "created", "updated").
			Values(1, "test_a", "test_b", 100.05, "2014-01-05", "2015-01-05").
			ToSql()
	}
}

//
// Update benchmark
//
func BenchmarkSquirrelUpdateSetColumns(b *testing.B) {
	for n := 0; n < b.N; n++ {
		squirrel.Update("mytable").
			Set("foo", 1).
			Set("bar", squirrel.Expr("COALESCE(bar, 0) + 1")).
			Set("c", 2).
			Where("id = ?", 9).
			Limit(10).
			Offset(20).
			ToSql()
	}
}

func BenchmarkSquirrelUpdateSetMap(b *testing.B) {
	for n := 0; n < b.N; n++ {
		squirrel.Update("mytable").
			SetMap(map[string]interface{}{"b": 1, "c": 2, "bar": squirrel.Expr("COALESCE(bar, 0) + 1")}).
			Where("id = ?", 9).
			Limit(10).
			Offset(20).
			ToSql()
	}
}

//
// Delete benchmark
//
func BenchmarkSquirrelDelete(b *testing.B) {
	for n := 0; n < b.N; n++ {
		squirrel.Delete("test_table").
			Where("b = ?", 1).
			OrderBy("c").
			Limit(2).
			Offset(3).
			ToSql()
	}
}
