package flow

import "go-flow/eventbus"

const (
	TriggerEvent        eventbus.Topic = "event.flow.trigger"
	InitFinishEvent     eventbus.Topic = "event.flow.init.finish"
	InitFailedEvent     eventbus.Topic = "event.flow.init.failed"
	ExecuteSucceedEvent eventbus.Topic = "event.flow.execute.succeed"
	ExecuteFailedEvent  eventbus.Topic = "event.flow.execute.failed"
	ExecutePauseEvent   eventbus.Topic = "event.flow.execute.pause"
	ExecuteResumeEvent  eventbus.Topic = "event.flow.execute.resume"
	ExecuteCancelEvent  eventbus.Topic = "event.flow.execute.cancel"
)
