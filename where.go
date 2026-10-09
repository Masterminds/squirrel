package squirrel

import (
	"fmt"
	"strings"
)

type wherePart part

func newWherePart(pred interface{}, args ...interface{}) Sqlizer {
	return &wherePart{pred: pred, args: args}
}

func (p wherePart) ToSql() (sql string, args []interface{}, err error) {
	switch pred := p.pred.(type) {
	case nil:
		// no-op
	case rawSqlizer:
		sql, args, err = pred.toSqlRaw()
	case Sqlizer:
		sql, args, err = pred.ToSql()
	case map[string]interface{}:
		sql, args, err = Eq(pred).ToSql()
	case string:
		sql = pred
		args = p.args
	default:
		err = fmt.Errorf("expected string-keyed map or string, not %T", pred)
	}
	if err != nil || len(sql) == 0 {
		return
	}
	if hasTopLevelOr(sql) && !isEnclosedInParen(sql) {
		sql = fmt.Sprintf("(%s)", sql)
	}
	return
}

type parenPart part

// Paren wraps a predicate or expression in parentheses if not already enclosed.
// Ex:
//     .Where(Paren("a = ? OR b = ?", 1, 2))
//     .Where(Paren(Eq{"x": 1, "y": 2}))
func Paren(pred interface{}, args ...interface{}) Sqlizer {
	return &parenPart{pred: pred, args: args}
}

// ParenExpr builds an expression wrapped in parentheses.
// Ex:
//     ParenExpr("a = ? OR b = ?", 1, 2)
func ParenExpr(sql string, args ...interface{}) Sqlizer {
	return Paren(sql, args...)
}

func (p parenPart) ToSql() (sql string, args []interface{}, err error) {
	return p.toSql(false)
}

func (p parenPart) toSqlRaw() (sql string, args []interface{}, err error) {
	return p.toSql(true)
}

func (p parenPart) toSql(raw bool) (sql string, args []interface{}, err error) {
	switch pred := p.pred.(type) {
	case nil:
		// no-op
	case rawSqlizer:
		if raw {
			sql, args, err = pred.toSqlRaw()
		} else if s, ok := pred.(Sqlizer); ok {
			sql, args, err = s.ToSql()
		} else {
			sql, args, err = pred.toSqlRaw()
		}
	case Sqlizer:
		sql, args, err = nestedToSql(pred)
	case map[string]interface{}:
		sql, args, err = Eq(pred).ToSql()
	case string:
		sql = pred
		args = p.args
	default:
		err = fmt.Errorf("expected string-keyed map, string, or Sqlizer, not %T", pred)
	}
	if err != nil || len(sql) == 0 {
		return
	}
	if !isEnclosedInParen(sql) {
		sql = fmt.Sprintf("(%s)", sql)
	}
	return
}

func hasTopLevelOr(sql string) bool {
	sql = strings.TrimSpace(sql)
	if len(sql) < 2 {
		return false
	}

	var (
		inSingleQuote  bool
		inDoubleQuote  bool
		inBacktick     bool
		inLineComment  bool
		inBlockComment bool
		parenDepth     int
	)

	n := len(sql)
	for i := 0; i < n; i++ {
		c := sql[i]

		if inLineComment {
			if c == '\n' {
				inLineComment = false
			}
			continue
		}

		if inBlockComment {
			if c == '*' && i+1 < n && sql[i+1] == '/' {
				inBlockComment = false
				i++
			}
			continue
		}

		if inSingleQuote {
			if c == '\'' {
				if i+1 < n && sql[i+1] == '\'' {
					i++
				} else {
					inSingleQuote = false
				}
			}
			continue
		}

		if inDoubleQuote {
			if c == '"' {
				inDoubleQuote = false
			}
			continue
		}

		if inBacktick {
			if c == '`' {
				inBacktick = false
			}
			continue
		}

		if c == '-' && i+1 < n && sql[i+1] == '-' {
			inLineComment = true
			i++
			continue
		}
		if c == '/' && i+1 < n && sql[i+1] == '*' {
			inBlockComment = true
			i++
			continue
		}

		if c == '\'' {
			inSingleQuote = true
			continue
		}
		if c == '"' {
			inDoubleQuote = true
			continue
		}
		if c == '`' {
			inBacktick = true
			continue
		}

		if c == '(' {
			parenDepth++
			continue
		}
		if c == ')' {
			if parenDepth > 0 {
				parenDepth--
			}
			continue
		}

		if parenDepth == 0 {
			if (c == 'o' || c == 'O') && i+1 < n && (sql[i+1] == 'r' || sql[i+1] == 'R') {
				prevOk := (i == 0) || !isWordChar(sql[i-1])
				nextOk := (i+2 == n) || !isWordChar(sql[i+2])
				if prevOk && nextOk {
					return true
				}
			}
		}
	}
	return false
}

func isEnclosedInParen(sql string) bool {
	sql = strings.TrimSpace(sql)
	n := len(sql)
	if n < 2 || sql[0] != '(' || sql[n-1] != ')' {
		return false
	}

	var (
		inSingleQuote  bool
		inDoubleQuote  bool
		inBacktick     bool
		inLineComment  bool
		inBlockComment bool
		parenDepth     int
	)

	for i := 0; i < n; i++ {
		c := sql[i]

		if inLineComment {
			if c == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment {
			if c == '*' && i+1 < n && sql[i+1] == '/' {
				inBlockComment = false
				i++
			}
			continue
		}
		if inSingleQuote {
			if c == '\'' {
				if i+1 < n && sql[i+1] == '\'' {
					i++
				} else {
					inSingleQuote = false
				}
			}
			continue
		}
		if inDoubleQuote {
			if c == '"' {
				inDoubleQuote = false
			}
			continue
		}
		if inBacktick {
			if c == '`' {
				inBacktick = false
			}
			continue
		}

		if c == '-' && i+1 < n && sql[i+1] == '-' {
			inLineComment = true
			i++
			continue
		}
		if c == '/' && i+1 < n && sql[i+1] == '*' {
			inBlockComment = true
			i++
			continue
		}

		if c == '\'' {
			inSingleQuote = true
			continue
		}
		if c == '"' {
			inDoubleQuote = true
			continue
		}
		if c == '`' {
			inBacktick = true
			continue
		}

		if c == '(' {
			parenDepth++
		} else if c == ')' {
			parenDepth--
			if parenDepth == 0 && i < n-1 {
				return false
			}
		}
	}
	return parenDepth == 0
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '_' || c == '$' || c == '@'
}
