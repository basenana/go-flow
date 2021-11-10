package plugin

import (
	"github.com/zwwhdls/go-flow/flow"
)

type SingleFlow struct {
}

func (s SingleFlow) NextBatch(ctx flow.FlowContext) ([]flow.Task, error) {
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

func (s SingleFlow) Setup(ctx flow.FlowContext) {
	panic("implement me")
}

func (s SingleFlow) Teardown(ctx flow.FlowContext) {
	panic("implement me")
}

var _ flow.Flow = &SingleFlow{}

type SingleFlowBuilder struct {
}

var _ FlowBuilder = SingleFlowBuilder{}
