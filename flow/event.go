package flow

import (
	"fmt"
	"github.com/zwwhdls/go-flow/eventbus"
	"github.com/zwwhdls/go-flow/fsm"
	"strings"
)

const flowTopic = "event.flow"

const (
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
)

func EventTopic(fid FID) eventbus.Topic {
	return eventbus.Topic(fmt.Sprintf("%s.%s", flowTopic, fid))
}

func IsTaskEvent(eventType fsm.EventType) bool {
	return strings.HasPrefix(string(eventType), "task.")
}
