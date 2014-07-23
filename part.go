package squirrel

import (
	"io"
	"fmt"
)

type part struct {
	pred interface{}
	args []interface{}
}

type sqlPart interface {
	ToSql() (sql string, args []interface{}, err error)
}

func (p part) ToSql() (sql string, args []interface{}, err error) {
	switch pred := p.pred.(type) {
	case nil:
		// no-op
	case string:
		sql = pred
		args = p.args
	default:
		err = fmt.Errorf("expected string, not %T", pred)
	}
	return
}

func appendToSql(parts []sqlPart, w io.Writer, sep string, args []interface{}) ([]interface{}, error) {
	for i, p := range parts {
		partSql, partArgs, err := p.ToSql()
		if err != nil {
			return nil, err
		} else if len(partSql) == 0 {
			continue
		}

		if i > 0 {
			_, err := io.WriteString(w, sep)
			if err != nil {
				return nil, err
			}
		}

		_, err = io.WriteString(w, partSql)
		if err != nil {
			return nil, err
		}
		args = append(args, partArgs...)
	}
	return args, nil
}
