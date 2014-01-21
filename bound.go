package squirrel

type BoundBuilder interface {
	Select(columns ...string) selectBuilder
}

type boundBuilder struct {
	runner Runner
}

func NewBoundBuilder(runner Runner) BoundBuilder {
	return &boundBuilder{runner}
}

func (b *boundBuilder) Select(columns ...string) selectBuilder {
	return selectWith(b.runner, columns...)
}
