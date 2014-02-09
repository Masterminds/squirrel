package squirrel

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.True(t, stub.Scanned, "row was not scanned")
	assert.NoError(t, err)
}

func TestRowScanErr(t *testing.T) {
	stub := &RowStub{}
	rowErr := fmt.Errorf("scan err")
	row := &Row{RowScanner: stub, err: rowErr}
	err := row.Scan()
	assert.False(t, stub.Scanned, "row was scanned")
	assert.Equal(t, rowErr, err)
}
