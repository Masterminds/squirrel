package squirrel

import "github.com/lann/builder"

type statementBuilder builder.Builder

func (b statementBuilder) Select(columns ...string) SelectBuilder {
	return SelectBuilder(b).Columns(columns...)
}

func (b statementBuilder) RunWith(runner Runner) statementBuilder {
	return builder.Set(b, "RunWith", runner).(statementBuilder)
}

// StatementBuilder is a parent builder for other statement builders.
//
// Currently StatementBuilder has only two methods: RunWith, which has the same
// semantics as SelectBuilder.RunWith, and Select, which returns a SelectBuilder
// with RunWith already set.
var StatementBuilder = statementBuilder(builder.EmptyBuilder)

// Select returns a new SelectBuilder, optionally setting some result columns.
//
// See SelectBuilder.Columns.
func Select(columns ...string) SelectBuilder {
	return StatementBuilder.Select(columns...)
}
