package ext

import (
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
)

type Task struct {
	name    flow.TName
	status  fsm.Status
	message string

	setupFunc    func(ctx *flow.Context) error
	doFunc       func(ctx *flow.Context) error
	teardownFunc func(ctx *flow.Context)
}

var _ flow.Task = &Task{}

func (t Task) GetStatus() fsm.Status {
	return t.status
}

func (t *Task) SetStatus(status fsm.Status) {
	t.status = status
}

func (t Task) GetMessage() string {
	return t.message
}

func (t *Task) SetMessage(msg string) {
	t.message = msg
}

func (t Task) Name() flow.TName {
	return t.name
}

func (t Task) Setup(ctx *flow.Context) error {
	return t.setupFunc(ctx)
}

func (t Task) Do(ctx *flow.Context) error {
	return t.doFunc(ctx)
}

func (t Task) Teardown(ctx *flow.Context) {
	t.teardownFunc(ctx)
}

func BuildSampleTask(uniqueName string, spec TaskSpec) flow.Task {
	return &Task{
		name:   flow.TName(uniqueName),
		status: flow.CreatingStatus,

		setupFunc:    spec.Setup,
		doFunc:       spec.Do,
		teardownFunc: spec.Teardown,
	}
}

type TaskSpec struct {
	Setup    func(ctx *flow.Context) error
	Do       func(ctx *flow.Context) error
	Teardown func(ctx *flow.Context)
}
