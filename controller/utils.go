package controller

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/zwwhdls/go-flow/eventbus"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
)

func waitTaskRunningOrClose(topic eventbus.Topic, task flow.Task) fsm.Status {
	if task.GetStatus() == flow.RunningStatus {
		return flow.RunningStatus
	}

	var (
		waitTask = make(chan struct{})
		lID      = fmt.Sprintf("wait-task-%s-running-%s", task.Name(), uuid.New().String())
	)
	eventbus.Register(topic, eventbus.NewBlockListener(lID, func(obj interface{}, args ...interface{}) error {
		evt, ok := obj.(fsm.Event)
		if !ok {
			return fmt.Errorf("topic %s get non-event obj", topic)
		}

		if !flow.IsTaskEvent(evt.Type) {
			return nil
		}

		t := evt.Obj.(flow.Task)
		if t.Name() != task.Name() {
			return nil
		}

		switch evt.Type {
		case flow.TaskTriggerEvent, flow.TaskExecutePauseEvent:
			return nil
		default:
			close(waitTask)
		}

		return nil
	}))
	<-waitTask
	eventbus.Unregister(lID)
	return task.GetStatus()
}
