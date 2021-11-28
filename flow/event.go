package flow

import (
	"fmt"
	"github.com/zwwhdls/go-flow/eventbus"
)

type EventTopicTpl string

const (
	TriggerEventTopicTpl        EventTopicTpl = "event.flow.%s.trigger"
	ReTriggerEventTopicTpl      EventTopicTpl = "event.flow.%s.retrigger"
	InitFinishEventTopicTpl     EventTopicTpl = "event.flow.%s.init.finish"
	ExecuteSucceedEventTopicTpl EventTopicTpl = "event.flow.%s.execute.succeed"
	ExecuteFailedEventTopicTpl  EventTopicTpl = "event.flow.%s.execute.failed"
	ExecuteRetryEventTopicTpl   EventTopicTpl = "event.flow.%s.execute.retry"
	ExecutePauseEventTopicTpl   EventTopicTpl = "event.flow.%s.execute.pause"
	ExecuteResumeEventTopicTpl  EventTopicTpl = "event.flow.%s.execute.resume"
	ExecuteCancelEventTopicTpl  EventTopicTpl = "event.flow.%s.execute.cancel"

	TaskTriggerEventTopicTpl        EventTopicTpl = "event.flow.%s.task.%s.trigger"
	TaskReTriggerEventTopicTpl      EventTopicTpl = "event.flow.%s.task.%s.retrigger"
	TaskInitFinishEventTopicTpl     EventTopicTpl = "event.flow.%s.task.%s.init.finish"
	TaskExecuteSucceedEventTopicTpl EventTopicTpl = "event.flow.%s.task.%s.execute.succeed"
	TaskExecuteFailedEventTopicTpl  EventTopicTpl = "event.flow.%s.task.%s.execute.failed"
	TaskExecuteRetryEventTopicTpl   EventTopicTpl = "event.flow.%s.task.%s.execute.retry"
	TaskExecutePauseEventTopicTpl   EventTopicTpl = "event.flow.%s.task.%s.execute.pause"
	TaskExecuteResumeEventTopicTpl  EventTopicTpl = "event.flow.%s.task.%s.execute.resume"
	TaskExecuteCancelEventTopicTpl  EventTopicTpl = "event.flow.%s.task.%s.execute.cancel"
)

func GetFlowTopic(tpl EventTopicTpl, fid FID) eventbus.Topic {
	return eventbus.Topic(fmt.Sprintf(string(tpl), fid))
}

func GetTaskTopic(tpl EventTopicTpl, fid FID, tName TName) eventbus.Topic {
	return eventbus.Topic(fmt.Sprintf(string(tpl), fid, tName))
}
