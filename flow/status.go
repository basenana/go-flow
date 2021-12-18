package flow

import "github.com/zwwhdls/go-flow/fsm"

const (
	CreatingStatus     fsm.Status = "creating"
	InitializingStatus            = "initializing"
	RunningStatus                 = "running"
	SucceedStatus                 = "succeed"
	FailedStatus                  = "failed"
	PausedStatus                  = "paused"
	CanceledStatus                = "canceled"

	PendingStatus          = "pending"
	TaskInitializingStatus = "taskInitializing"
	TaskRunningStatus      = "taskRunning"
	TaskSucceedStatus      = "taskSucceed"
	TaskFailedStatus       = "taskFailed"
)
