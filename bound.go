package squirrel

type boundBuilder struct {
	runner Runner
}

func NewBoundBuilder(runner Runner) *boundBuilder {
	return &boundBuilder{runner}
}

func (b *boundBuilder) Select(columns ...string) selectBuilder {
	return selectWith(b.runner, columns...)
}
