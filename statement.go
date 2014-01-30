package squirrel

import "github.com/lann/builder"

// StatementBuilderType is the type of StatementBuilder.
type StatementBuilderType builder.Builder

// Select returns a SelectBuilder for this StatementBuilderType.
func (b StatementBuilderType) Select(columns ...string) SelectBuilder {
	return SelectBuilder(b).Columns(columns...)
}

// RunWith sets the RunWith field for any child builders.
func (b StatementBuilderType) RunWith(runner Runner) StatementBuilderType {
	return builder.Set(b, "RunWith", runner).(StatementBuilderType)
}

// StatementBuilder is a parent builder for other builders, e.g. SelectBuilder.
var StatementBuilder = StatementBuilderType(builder.EmptyBuilder)

// Select returns a new SelectBuilder, optionally setting some result columns.
//
// See SelectBuilder.Columns.
func Select(columns ...string) SelectBuilder {
	return StatementBuilder.Select(columns...)
}
