package task

import "go-flow/eventbus"

const (
	TriggerEvent        eventbus.Topic = "event.task.trigger"
	InitFinishEvent     eventbus.Topic = "event.task.init.finish"
	InitFailedEvent     eventbus.Topic = "event.task.init.failed"
	ExecuteSucceedEvent eventbus.Topic = "event.task.execute.succeed"
	ExecuteFailedEvent  eventbus.Topic = "event.task.execute.failed"
	ExecutePauseEvent   eventbus.Topic = "event.task.execute.pause"
	ExecuteResumeEvent  eventbus.Topic = "event.task.execute.resume"
	ExecuteCancelEvent  eventbus.Topic = "event.task.execute.cancel"
)
