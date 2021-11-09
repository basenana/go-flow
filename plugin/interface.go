package plugin

import (
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
)

type FlowBuilder interface {
	Build() flow.Flow
	GetFlowHook(flow.Flow) Hook
	GetTaskHook(flow.Flow, flow.Task) Hook
}

type Hook struct {
	WhenTrigger        fsm.Handler
	WhenInitFinish     fsm.Handler
	WhenExecuteSucceed fsm.Handler
	WhenExecuteFailed  fsm.Handler
	WhenExecutePause   fsm.Handler
	WhenExecuteResume  fsm.Handler
	WhenExecuteCancel  fsm.Handler

	WhenTaskTrigger        fsm.Handler
	WhenTaskInitFinish     fsm.Handler
	WhenTaskInitFailed     fsm.Handler
	WhenTaskExecuteSucceed fsm.Handler
	WhenTaskExecuteFailed  fsm.Handler
	WhenTaskExecutePause   fsm.Handler
	WhenTaskExecuteResume  fsm.Handler
	WhenTaskExecuteCancel  fsm.Handler
}
