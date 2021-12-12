package plugin

import (
	"github.com/zwwhdls/go-flow/flow"
)

type FlowBuilder interface {
	Build() flow.Flow
}

type HookType string
type Hooks map[HookType]Hook

const (
	WhenTrigger        HookType = "Trigger"
	WhenExecuteSucceed HookType = "Succeed"
	WhenExecuteFailed  HookType = "Failed"
	WhenExecutePause   HookType = "Pause"
	WhenExecuteResume  HookType = "Resume"
	WhenExecuteCancel  HookType = "Cancel"

	WhenTaskTrigger        HookType = "TaskTrigger"
	WhenTaskExecuteSucceed HookType = "TaskSucceed"
	WhenTaskExecuteFailed  HookType = "TaskFailed"
	WhenTaskExecutePause   HookType = "TaskPause"
	WhenTaskExecuteResume  HookType = "TaskResume"
	WhenTaskExecuteCancel  HookType = "TaskCancel"
)

type Hook func(ctx *flow.Context, f flow.Flow, t flow.Task) error
