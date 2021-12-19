package flow

import (
	"github.com/zwwhdls/go-flow/fsm"
)

type Flow interface {
	fsm.Stateful
	ID() FID
	GetHooks() Hooks
	Setup(ctx *Context) error
	Teardown(ctx *Context)
	NextBatch(ctx *Context) ([]Task, error)
	GetControlPolicy() ControlPolicy
}

type Task interface {
	fsm.Stateful
	Name() TName
	Setup(ctx *Context) error
	Do(ctx *Context) error
	Teardown(ctx *Context)
}

type ControlPolicy struct {
	FailedPolicy FailedPolicy
}

type (
	FID          string
	TName        string
	FailedPolicy string
	HookType     string
	Hooks        map[HookType]Hook
	Hook         func(ctx *Context, f Flow, t Task) error
)

const (
	CreatingStatus     fsm.Status = "creating"
	InitializingStatus            = "initializing"
	RunningStatus                 = "running"
	SucceedStatus                 = "succeed"
	FailedStatus                  = "failed"
	PausedStatus                  = "paused"
	CanceledStatus                = "canceled"

	TriggerEvent           fsm.EventType = "flow.execute.trigger"
	ExecuteFinishEvent                   = "flow.execute.finish"
	ExecuteErrorEvent                    = "flow.execute.error"
	ExecutePauseEvent                    = "flow.execute.pause"
	ExecuteResumeEvent                   = "flow.execute.resume"
	ExecuteCancelEvent                   = "flow.execute.cancel"
	TaskTriggerEvent                     = "task.execute.trigger"
	TaskExecutePauseEvent                = "task.execute.pause"
	TaskExecuteResumeEvent               = "task.execute.resume"
	TaskExecuteCancelEvent               = "task.execute.cancel"
	TaskExecuteFinishEvent               = "task.execute.finish"
	TaskExecuteErrorEvent                = "task.execute.error"

	PolicyFastFailed FailedPolicy = "fastFailed"
	PolicyPaused                  = "paused"

	WhenTrigger            HookType = "Trigger"
	WhenExecuteSucceed              = "Succeed"
	WhenExecuteFailed               = "Failed"
	WhenExecutePause                = "Pause"
	WhenExecuteResume               = "Resume"
	WhenExecuteCancel               = "Cancel"
	WhenTaskTrigger                 = "TaskTrigger"
	WhenTaskExecuteSucceed          = "TaskSucceed"
	WhenTaskExecuteFailed           = "TaskFailed"
	WhenTaskExecutePause            = "TaskPause"
	WhenTaskExecuteResume           = "TaskResume"
	WhenTaskExecuteCancel           = "TaskCancel"
)
