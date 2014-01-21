package squirrel

import (
	"database/sql"
	"fmt"
	"testing"
)

type DBStub struct {
	err              error

	LastPrepareSql   string
	PrepareCount     int

	LastExecSql      string
	LastExecArgs     []interface{}

	LastQuerySql     string
	LastQueryArgs    []interface{}

	LastQueryRowSql  string
	LastQueryRowArgs []interface{}
}

var StubError = fmt.Errorf("this is a stub; this is only a stub")

func (s *DBStub) Prepare(query string) (*sql.Stmt, error) {
	s.LastPrepareSql = query
	s.PrepareCount++
	return nil, nil
}

func (s *DBStub) Exec(query string, args ...interface{}) (sql.Result, error) {
	s.LastExecSql = query
	s.LastExecArgs = args
	return nil, nil
}

func (s *DBStub) Query(query string, args ...interface{}) (*sql.Rows, error) {
	s.LastQuerySql = query
	s.LastQueryArgs = args
	return nil, nil
}

func (s *DBStub) QueryRow(query string, args ...interface{}) RowScanner {
	s.LastQueryRowSql = query
	s.LastQueryRowArgs = args
	return &Row{RowScanner: &RowStub{}}
}

var sqlizer = Select("test")
var sqlStr = "SELECT test"

func TestExecWith(t *testing.T) {
	db := &DBStub{}
	ExecWith(db, sqlizer)
	if db.LastExecSql != sqlStr {
		t.Errorf("expected %v, got %v", sqlStr, db.LastExecSql)
	}
}

func TestQueryWith(t *testing.T) {
	db := &DBStub{}
	QueryWith(db, sqlizer)
	if db.LastQuerySql != sqlStr {
		t.Errorf("expected %v, got %v", sqlStr, db.LastQuerySql)
	}
}

func TestQueryRowWith(t *testing.T) {
	db := &DBStub{}
	QueryRowWith(db, sqlizer)
	if db.LastQueryRowSql != sqlStr {
		t.Errorf("expected %v, got %v", sqlStr, db.LastQueryRowSql)
	}
}

func TestWithToSqlErr(t *testing.T) {
	db := &DBStub{}
	sqlizer := Select()

	_, err := ExecWith(db, sqlizer)
	if err == nil {
		t.Error("expected error, got none")
	}

	_, err = QueryWith(db, sqlizer)
	if err == nil {
		t.Error("expected error, got none")
	}

	err = QueryRowWith(db, sqlizer).Scan()
	if err == nil {
		t.Error("expected error, got none")
	}
}
