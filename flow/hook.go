package flow

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

type Hook func(ctx Context, f Flow, t Task) error
