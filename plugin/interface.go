package plugin

import (
	"github.com/zwwhdls/go-flow/flow"
)

type FlowBuilder interface {
	Build() flow.Flow
	GetFlowHook(flow.Flow) Hook
	GetTaskHook(flow.Flow, flow.Task) Hook
}

type Hook struct {
	WhenTrigger        FlowHook
	WhenInitFinish     FlowHook
	WhenExecuteSucceed FlowHook
	WhenExecuteFailed  FlowHook
	WhenExecutePause   FlowHook
	WhenExecuteResume  FlowHook
	WhenExecuteCancel  FlowHook

	WhenTaskTrigger        TaskHook
	WhenTaskInitFinish     TaskHook
	WhenTaskInitFailed     TaskHook
	WhenTaskExecuteSucceed TaskHook
	WhenTaskExecuteFailed  TaskHook
	WhenTaskExecutePause   TaskHook
	WhenTaskExecuteResume  TaskHook
	WhenTaskExecuteCancel  TaskHook
}

type FlowHook func(ctx *flow.FlowContext, f flow.Flow) error
type TaskHook func(ctx *flow.TaskContext, t flow.Task) error
