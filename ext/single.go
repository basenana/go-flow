package ext

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/storage"
)

type SingleFlow struct {
	*basic
	tasks []flow.Task
}

func (s SingleFlow) Type() flow.FType {
	return "Single"
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

func NewSingleFlow(s storage.Interface, tasks []flow.Task, policy flow.ControlPolicy) error {
	f := &SingleFlow{
		basic: &basic{
			id:     flow.FID(fmt.Sprintf("single-flow-%s", uuid.New().String())),
			status: flow.CreatingStatus,
			policy: policy,
		},
		tasks: tasks,
	}
	return s.SaveFlow(f)
}
