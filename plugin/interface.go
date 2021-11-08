package plugin

import "go-flow/task"

type FlowBuilder interface {
	WhenTrigger() error
	WhenInitFinish() error
	WhenInitFailed() error
	WhenExecuteSucceed() error
	WhenExecuteFailed() error
	WhenExecutePause() error
	WhenExecuteResume() error
	WhenExecuteCancel() error

	WhenTaskTrigger(taskName task.TName) error
	WhenTaskInitFinish(taskName task.TName) error
	WhenTaskInitFailed(taskName task.TName) error
	WhenTaskExecuteSucceed(taskName task.TName) error
	WhenTaskExecuteFailed(taskName task.TName) error
	WhenTaskExecutePause(taskName task.TName) error
	WhenTaskExecuteResume(taskName task.TName) error
	WhenTaskExecuteCancel(taskName task.TName) error
}
