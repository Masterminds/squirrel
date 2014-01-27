package squirrel

import "github.com/lann/builder"

type statementBuilder builder.Builder

func (b statementBuilder) Select(columns ...string) selectBuilder {
	return selectBuilder(b).Columns(columns...)
}

func (b statementBuilder) RunWith(runner Runner) statementBuilder {
	return builder.Set(b, "RunWith", runner).(statementBuilder)
}

var StatementBuilder = statementBuilder(builder.EmptyBuilder)

func Select(columns ...string) selectBuilder {
	return StatementBuilder.Select(columns...)
}
