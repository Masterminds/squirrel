package squirrel

import (
	"fmt"
	"testing"
)

type RowStub struct {
	Scanned bool
}

func (r *RowStub) Scan(_ ...interface{}) error {
	r.Scanned = true
	return nil
}

func TestRowScan(t *testing.T) {
	stub := &RowStub{}
	row := &Row{RowScanner: stub}
	err := row.Scan()
	if !stub.Scanned {
		t.Error("row was not scanned")
	}
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
}

func TestRowScanErr(t *testing.T) {
	stub := &RowStub{}
	rowErr := fmt.Errorf("scan err")
	row := &Row{RowScanner: stub, err: rowErr}
	err := row.Scan()
	if stub.Scanned {
		t.Error("row was scanned")
	}
	if err != rowErr {
		t.Errorf("expected %v, got %v", rowErr, err)
	}
}
