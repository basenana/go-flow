package ext

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/zwwhdls/go-flow/flow"
)

type SingleFlow struct {
	*basic
	tasks []flow.Task
}

var _ flow.Flow = &SingleFlow{}

func (s SingleFlow) GetHooks() flow.Hooks {
	return map[flow.HookType]flow.Hook{}
}

func (s SingleFlow) Setup(ctx *flow.Context) error {
	return nil
}

func (s SingleFlow) Teardown(ctx *flow.Context) {
	return
}

func (s *SingleFlow) NextBatch(ctx *flow.Context) ([]flow.Task, error) {
	if len(s.tasks) == 0 {
		return nil, nil
	}

	t := s.tasks[0]
	s.tasks = s.tasks[1:]

	return []flow.Task{t}, nil
}

type SingleFlowBuilder struct {
	tasks  []flow.Task
	policy flow.ControlPolicy
}

var _ FlowBuilder = SingleFlowBuilder{}

func (s SingleFlowBuilder) Build() flow.Flow {
	return &SingleFlow{
		basic: &basic{
			id:     flow.FID(fmt.Sprintf("single-flow-%s", uuid.New().String())),
			status: flow.CreatingStatus,
			policy: s.policy,
		},
		tasks: s.tasks,
	}
}

func NewSingleFlowBuilder(tasks []flow.Task, policy flow.ControlPolicy) FlowBuilder {
	return SingleFlowBuilder{tasks: tasks, policy: policy}
}
