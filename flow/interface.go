/*
   Copyright 2022 Go-Flow Authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package flow

import (
	"github.com/basenana/go-flow/fsm"
)

type ITask interface {
	fsm.Stateful
	Name() string
	Setup(ctx *Context) error
	Do(ctx *Context) error
	Teardown(ctx *Context)
}

const (
	CreatingStatus     = "creating"
	InitializingStatus = "initializing"
	RunningStatus      = "running"
	SucceedStatus      = "succeed"
	FailedStatus       = "failed"
	ErrorStatus        = "error"
	PausedStatus       = "paused"
	CanceledStatus     = "canceled"

	TriggerEvent           = "flow.execute.trigger"
	ExecuteFinishEvent     = "flow.execute.finish"
	ExecuteErrorEvent      = "flow.execute.error"
	ExecutePauseEvent      = "flow.execute.pause"
	ExecuteResumeEvent     = "flow.execute.resume"
	ExecuteCancelEvent     = "flow.execute.cancel"
	TaskTriggerEvent       = "task.execute.trigger"
	TaskExecuteCancelEvent = "task.execute.cancel"
	TaskExecuteFinishEvent = "task.execute.finish"
	TaskExecuteErrorEvent  = "task.execute.error"

	PolicyFastFailed = "fastFailed"
	PolicyPaused     = "paused"

	WhenTrigger            = "Trigger"
	WhenExecuteSucceed     = "Succeed"
	WhenExecuteFailed      = "Failed"
	WhenExecutePause       = "Pause"
	WhenExecuteResume      = "Resume"
	WhenExecuteCancel      = "Cancel"
	WhenTaskTrigger        = "TaskTrigger"
	WhenTaskExecuteSucceed = "TaskSucceed"
	WhenTaskExecuteFailed  = "TaskFailed"
	WhenTaskExecutePause   = "TaskPause"
	WhenTaskExecuteResume  = "TaskResume"
	WhenTaskExecuteCancel  = "TaskCancel"
)
