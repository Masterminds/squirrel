// +build integration

package squirrel

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

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
	sqrl StatementBuilderType
)

func TestMain(m *testing.M) {
	var driver, dataSource string
	flag.StringVar(&driver, "driver", "", "integration database driver")
	flag.StringVar(&dataSource, "dataSource", "", "integration database data source")
	flag.Parse()

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

	sqrl = StatementBuilder.RunWith(db)

	if driver == "postgres" {
		sqrl = sqrl.PlaceholderFormat(Dollar)
	}

	os.Exit(m.Run())
}

func assertVals(t *testing.T, s SelectBuilder, expected ...string) {
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
		sqrl.Select("v").From("squirrel_integration"),
		"foo", "bar", "foo", "baz")
}

func TestEq(t *testing.T) {
	s := sqrl.Select("v").From("squirrel_integration")
	assertVals(t, s.Where(Eq{"k": 4}), "baz")
	assertVals(t, s.Where(NotEq{"k": 2}), "foo", "bar", "baz")
	assertVals(t, s.Where(Eq{"k": []int{1,4}}), "foo", "baz")
	assertVals(t, s.Where(NotEq{"k": []int{1,4}}), "bar", "foo")
	assertVals(t, s.Where(Eq{"k": nil}))
	assertVals(t, s.Where(NotEq{"k": nil}), "foo", "bar", "foo", "baz")
	assertVals(t, s.Where(Eq{"k": []int{}}))
	assertVals(t, s.Where(NotEq{"k": []int{}}), "foo", "bar", "foo", "baz")
}

func TestIneq(t *testing.T) {
	s := sqrl.Select("v").From("squirrel_integration")
	assertVals(t, s.Where(Lt{"k": 3}), "foo", "foo")
	assertVals(t, s.Where(Gt{"k": 3}), "baz")
}

func TestConj(t *testing.T) {
	s := sqrl.Select("v").From("squirrel_integration")
	assertVals(t, s.Where(And{Gt{"k": 1}, Lt{"k": 4}}), "bar", "foo")
	assertVals(t, s.Where(Or{Gt{"k": 3}, Lt{"k": 2}}), "foo", "baz")
}

func TestContext(t *testing.T) {
	s := sqrl.Select("v").From("squirrel_integration")
	ctx := context.Background()
	_, err := s.QueryContext(ctx)
	assert.NoError(t, err)
}
