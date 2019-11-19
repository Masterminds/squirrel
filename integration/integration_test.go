package integration

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

  sqrl "github.com/Masterminds/squirrel"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const (
	testSchema = `
		CREATE TABLE squirrel_integration ( k INT, v TEXT )`
	testData = `
		INSERT INTO squirrel_integration VALUES
			(1, 'foo'),
			(3, 'bar'),
			(2, 'foo'),
			(4, 'baz')
		`
)

var (
	sb sqrl.StatementBuilderType
)

func TestMain(m *testing.M) {
	var driver, dataSource string
	flag.StringVar(&driver, "driver", "", "integration database driver")
	flag.StringVar(&dataSource, "dataSource", "", "integration database data source")
	flag.Parse()

  if driver == "" {
    driver = "sqlite3"
  }

	if driver == "sqlite3" && dataSource == "" {
		dataSource = ":memory:"
	}

	db, err := sql.Open(driver, dataSource)
	if err != nil {
		fmt.Printf("error opening database: %v\n", err)
		os.Exit(-1)
	}

	_, err = db.Exec(testSchema)
	if err != nil {
		fmt.Printf("error creating test schema: %v\n", err)
		os.Exit(-2)
	}

	defer func() {
		_, err = db.Exec("DROP TABLE squirrel_integration")
		fmt.Printf("error removing test schema: %v\n", err)
	}()

	_, err = db.Exec(testData)
	if err != nil {
		fmt.Printf("error inserting test data: %v\n", err)
		os.Exit(-3)
	}

	sb = sqrl.StatementBuilder.RunWith(db)

	if driver == "postgres" {
		sb = sb.PlaceholderFormat(sqrl.Dollar)
	}

	os.Exit(m.Run())
}

func assertVals(t *testing.T, s sqrl.SelectBuilder, expected ...string) {
	rows, err := s.Query()
	assert.NoError(t, err)
	defer rows.Close()

	vals := make([]string, len(expected))
	for i := range vals {
		assert.True(t, rows.Next())
		assert.NoError(t, rows.Scan(&vals[i]))
	}
	assert.False(t, rows.Next())

	if expected != nil {
		assert.Equal(t, expected, vals)
	}
}

func TestSimpleSelect(t *testing.T) {
	assertVals(
		t,
		sb.Select("v").From("squirrel_integration"),
		"foo", "bar", "foo", "baz")
}

func TestEq(t *testing.T) {
	s := sb.Select("v").From("squirrel_integration")
	assertVals(t, s.Where(sqrl.Eq{"k": 4}), "baz")
	assertVals(t, s.Where(sqrl.NotEq{"k": 2}), "foo", "bar", "baz")
	assertVals(t, s.Where(sqrl.Eq{"k": []int{1, 4}}), "foo", "baz")
	assertVals(t, s.Where(sqrl.NotEq{"k": []int{1, 4}}), "bar", "foo")
	assertVals(t, s.Where(sqrl.Eq{"k": nil}))
	assertVals(t, s.Where(sqrl.NotEq{"k": nil}), "foo", "bar", "foo", "baz")
	assertVals(t, s.Where(sqrl.Eq{"k": []int{}}))
	assertVals(t, s.Where(sqrl.NotEq{"k": []int{}}), "foo", "bar", "foo", "baz")
}

func TestIneq(t *testing.T) {
	s := sb.Select("v").From("squirrel_integration")
	assertVals(t, s.Where(sqrl.Lt{"k": 3}), "foo", "foo")
	assertVals(t, s.Where(sqrl.Gt{"k": 3}), "baz")
}

func TestConj(t *testing.T) {
	s := sb.Select("v").From("squirrel_integration")
	assertVals(t, s.Where(sqrl.And{sqrl.Gt{"k": 1}, sqrl.Lt{"k": 4}}), "bar", "foo")
	assertVals(t, s.Where(sqrl.Or{sqrl.Gt{"k": 3}, sqrl.Lt{"k": 2}}), "foo", "baz")
}

func TestContext(t *testing.T) {
	s := sb.Select("v").From("squirrel_integration")
	ctx := context.Background()
	_, err := s.QueryContext(ctx)
	assert.NoError(t, err)
}
