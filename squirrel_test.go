package squirrel

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DBStub struct {
	err error

	LastPrepareSQL string
	PrepareCount   int

	LastExecSQL  string
	LastExecArgs []interface{}

	LastQuerySQL  string
	LastQueryArgs []interface{}

	LastQueryRowSQL  string
	LastQueryRowArgs []interface{}
}

var ErrStubError = fmt.Errorf("this is a stub; this is only a stub")

func (s *DBStub) Prepare(query string) (*sql.Stmt, error) {
	s.LastPrepareSQL = query
	s.PrepareCount++
	return nil, nil
}

func (s *DBStub) Exec(query string, args ...interface{}) (sql.Result, error) {
	s.LastExecSQL = query
	s.LastExecArgs = args
	return nil, nil
}

func (s *DBStub) Query(query string, args ...interface{}) (*sql.Rows, error) {
	s.LastQuerySQL = query
	s.LastQueryArgs = args
	return nil, nil
}

func (s *DBStub) QueryRow(query string, args ...interface{}) RowScanner {
	s.LastQueryRowSQL = query
	s.LastQueryRowArgs = args
	return &Row{RowScanner: &RowStub{}}
}

var sqlizer = Select("test")
var sqlStr = "SELECT test"

func TestExecWith(t *testing.T) {
	db := &DBStub{}
	ExecWith(db, sqlizer)
	assert.Equal(t, sqlStr, db.LastExecSQL)
}

func TestQueryWith(t *testing.T) {
	db := &DBStub{}
	QueryWith(db, sqlizer)
	assert.Equal(t, sqlStr, db.LastQuerySQL)
}

func TestQueryRowWith(t *testing.T) {
	db := &DBStub{}
	QueryRowWith(db, sqlizer)
	assert.Equal(t, sqlStr, db.LastQueryRowSQL)
}

func TestWithToSQLErr(t *testing.T) {
	db := &DBStub{}
	sqlizer := Select()

	_, err := ExecWith(db, sqlizer)
	assert.Error(t, err)

	_, err = QueryWith(db, sqlizer)
	assert.Error(t, err)

	err = QueryRowWith(db, sqlizer).Scan()
	assert.Error(t, err)
}

func TestDebugSqlizer(t *testing.T) {
	sqlizer := Expr("x = ? AND y = ? AND z = '??'", 1, "text")
	expectedDebug := "x = '1' AND y = 'text' AND z = '?'"
	assert.Equal(t, expectedDebug, DebugSqlizer(sqlizer))
}

func TestDebugSqlizerErrors(t *testing.T) {
	errorMsg := DebugSqlizer(Expr("x = ?", 1, 2)) // Not enough placeholders
	assert.True(t, strings.HasPrefix(errorMsg, "[DebugSqlizer error: "))

	errorMsg = DebugSqlizer(Expr("x = ? AND y = ?", 1)) // Too many placeholders
	assert.True(t, strings.HasPrefix(errorMsg, "[DebugSqlizer error: "))

	errorMsg = DebugSqlizer(Lt{"x": nil}) // Cannot use nil values with Lt
	assert.True(t, strings.HasPrefix(errorMsg, "[ToSQL error: "))
}
