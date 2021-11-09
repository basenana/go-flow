package flow

import "github.com/zwwhdls/go-flow/eventbus"

const (
	TriggerEvent        eventbus.Topic = "event.flow.trigger"
	InitFinishEvent     eventbus.Topic = "event.flow.init.finish"
	ExecuteSucceedEvent eventbus.Topic = "event.flow.execute.succeed"
	ExecuteFailedEvent  eventbus.Topic = "event.flow.execute.failed"
	ExecutePauseEvent   eventbus.Topic = "event.flow.execute.pause"
	ExecuteResumeEvent  eventbus.Topic = "event.flow.execute.resume"
	ExecuteCancelEvent  eventbus.Topic = "event.flow.execute.cancel"

	TaskTriggerEvent        eventbus.Topic = "event.task.trigger"
	TaskInitFinishEvent     eventbus.Topic = "event.task.init.finish"
	TaskExecuteSucceedEvent eventbus.Topic = "event.task.execute.succeed"
	TaskExecuteFailedEvent  eventbus.Topic = "event.task.execute.failed"
	TaskExecutePauseEvent   eventbus.Topic = "event.task.execute.pause"
	TaskExecuteResumeEvent  eventbus.Topic = "event.task.execute.resume"
	TaskExecuteCancelEvent  eventbus.Topic = "event.task.execute.cancel"
)
