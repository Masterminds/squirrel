package squirrel

import "errors"

var (
	ErrNoTable  = errors.New("statement must specify a table")
	ErrNoValues = errors.New("statement must have at least one set of values or select clause")
)
