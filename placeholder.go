package squirrel

import (
	"bytes"
	"fmt"
	"strings"
)

// PlaceholderFormat is the interface that wraps the ReplacePlaceholders method.
//
// ReplacePlaceholders takes a SQL statement and replaces each question mark
// placeholder with a (possibly different) SQL placeholder.
type PlaceholderFormat interface {
	ReplacePlaceholders(sql string) (string, error)
}

var (
	// Question is a PlaceholderFormat instance that leaves placeholders as
	// question marks.
	Question = questionFormat{}

	// Dollar is a PlaceholderFormat instance that replaces placeholders with
	// dollar-prefixed positional placeholders (e.g. $1, $2, $3).
	Dollar   = dollarFormat{}
)

type questionFormat struct{}

func (_ questionFormat) ReplacePlaceholders(sql string) (string, error) {
	return sql, nil
}

type dollarFormat struct{}

func (_ dollarFormat) ReplacePlaceholders(sql string) (string, error) {
	buf := &bytes.Buffer{}
	for i := 1;; i++ {
		p := strings.Index(sql, "?")
		if p == -1 {
			break
		}

		buf.WriteString(sql[:p])
		fmt.Fprintf(buf, "$%d", i)
		sql = sql[p + 1:]
	}

	buf.WriteString(sql)
	return buf.String(), nil
}
