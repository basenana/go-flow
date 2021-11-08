package plugin

import (
	"go-flow/context"
	"go-flow/flow"
	"go-flow/task"
)

type SingleFlow struct {
}

func (s SingleFlow) NextBatch(ctx context.FlowContext) ([]task.Task, error) {
	panic("implement me")
}

func (s SingleFlow) ID() flow.FID {
	panic("implement me")
}

func (s SingleFlow) GetStatus() string {
	panic("implement me")
}

func (s SingleFlow) SetStatus(s2 string) error {
	panic("implement me")
}

func (s SingleFlow) Setup(ctx context.FlowContext) {
	panic("implement me")
}

func (s SingleFlow) Teardown(ctx context.FlowContext) {
	panic("implement me")
}

var _ flow.Flow = &SingleFlow{}

type SingleFlowBuilder struct {
}

var _ FlowBuilder = SingleFlowBuilder{}
