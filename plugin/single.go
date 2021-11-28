package plugin

import (
	"fmt"
	"github.com/zwwhdls/go-flow/flow"
	"time"
)

type SingleFlow struct {
	id     string
	status string
	tasks  []*SingleTask
}

func (s *SingleFlow) ID() flow.FID {
	return flow.FID(s.id)
}

func (s *SingleFlow) GetStatus() string {
	return s.status
}

func (s *SingleFlow) SetStatus(s2 string) error {
	s.status = s2
	return nil
}

func (s *SingleFlow) Setup(ctx *flow.FlowContext) error {
	fmt.Printf("flow %s init succeed.\n", s.id)
	return nil
}

func (s *SingleFlow) Teardown(ctx *flow.FlowContext) {
	fmt.Printf("flow %s tear down succeed.\n", s.id)
}

func (s *SingleFlow) NextBatch(ctx *flow.FlowContext) ([]flow.Task, error) {
	tasks := []flow.Task{}
	for _, t := range s.tasks {
		if t.GetStatus() != flow.TaskSucceedStatus {
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

var _ flow.Flow = &SingleFlow{}

type SingleTask struct {
	name   string
	status string
}

var _ flow.Task = &SingleTask{}

func (s *SingleTask) GetStatus() string {
	return s.status
}

func (s *SingleTask) SetStatus(s2 string) error {
	s.status = s2
	return nil
}

func (s *SingleTask) Name() flow.TName {
	return flow.TName(s.name)
}

func (s *SingleTask) Setup(ctx *flow.TaskContext) error {
	fmt.Printf("task %s init succeed.\n", s.name)
	return nil
}

func (s *SingleTask) Do(ctx *flow.TaskContext) {
	time.Sleep(1 * time.Second)
	fmt.Printf("task %s execute succeed.\n", s.name)
	ctx.Succeed()
}

func (s *SingleTask) Teardown(ctx *flow.TaskContext) {
	fmt.Printf("task %s tear down succeed.\n", s.name)
}

type SingleFlowBuilder struct {
}

func (s SingleFlowBuilder) Build() flow.Flow {
	f := SingleFlow{
		id: "f",
		tasks: []*SingleTask{
			{name: "t1"},
			{name: "t2"},
			{name: "t3"},
		},
	}
	return &f
}

func (s SingleFlowBuilder) GetFlowHook(f flow.Flow) Hook {
	return Hook{
		WhenTrigger: func(ctx *flow.FlowContext, f flow.Flow) error {
			fmt.Println("flow trigger")
			return nil
		},
		WhenInitFinish: nil,
		WhenExecuteSucceed: func(ctx *flow.FlowContext, f flow.Flow) error {
			fmt.Println("flow succeed")
			return nil
		},
		WhenExecuteFailed: func(ctx *flow.FlowContext, f flow.Flow) error {
			fmt.Println("flow failed")
			return nil
		},
		WhenExecutePause:  nil,
		WhenExecuteResume: nil,
		WhenExecuteCancel: nil,
	}
}

func (s SingleFlowBuilder) GetTaskHook(f flow.Flow, task flow.Task) Hook {
	return Hook{
		WhenTaskTrigger: func(ctx *flow.TaskContext, t flow.Task) error {
			fmt.Printf("task succeed")
			return nil
		},
		WhenTaskInitFinish:     nil,
		WhenTaskInitFailed:     nil,
		WhenTaskExecuteSucceed: nil,
		WhenTaskExecuteFailed:  nil,
		WhenTaskExecutePause:   nil,
		WhenTaskExecuteResume:  nil,
		WhenTaskExecuteCancel:  nil,
	}
}

var _ FlowBuilder = SingleFlowBuilder{}
