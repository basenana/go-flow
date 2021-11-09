package controller

import (
	"fmt"
	"github.com/zwwhdls/go-flow/context"
	"github.com/zwwhdls/go-flow/eventbus"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
)

func (c *FlowController) triggerTask(f *flowWarp, t flow.Task) {
	twarp := taskWarp{
		Task:    t,
		fid:     f.ID(),
		machine: initTaskFsm(f, t),
	}
	ctx := context.TaskContext{}
	eventbus.Publish(flow.TaskTriggerEvent, f)
	twarp.Do(ctx)
}

func initTaskFsm(f *flowWarp, t flow.Task) *fsm.FSM {
	hook := buildTaskHookWithDefault(f.builder.GetTaskHook(f.Flow, t))
	m := fsm.New(fsm.Option{
		Name:   fmt.Sprintf("task.%s.%s", f.ID(), t.Name()),
		Obj:    f,
		Logger: nil,
	})

	m.From([]string{flow.PendingStatus}).
		To(flow.InitializingStatus).
		When(flow.TaskTriggerEvent).
		Do(hook.WhenTaskTrigger)

	m.From([]string{flow.InitializingStatus}).
		To(flow.RunningStatus).
		When(flow.TaskInitFinishEvent).
		Do(hook.WhenTaskInitFinish)

	m.From([]string{flow.RunningStatus}).
		To(flow.SucceedStatus).
		When(flow.TaskExecuteSucceedEvent).
		Do(hook.WhenTaskExecuteSucceed)

	m.From([]string{flow.InitializingStatus, flow.RunningStatus}).
		To(flow.FailedStatus).
		When(flow.TaskExecuteFailedEvent).
		Do(hook.WhenExecuteFailed)

	m.From([]string{flow.PausedStatus, flow.InitializingStatus, flow.RunningStatus}).
		To(flow.CanceledStatus).
		When(flow.TaskExecuteCancelEvent).
		Do(hook.WhenTaskExecuteCancel)

	m.From([]string{flow.RunningStatus}).
		To(flow.PausedStatus).
		When(flow.TaskExecutePauseEvent).
		Do(hook.WhenTaskExecutePause)

	m.From([]string{flow.PausedStatus}).
		To(flow.RunningStatus).
		When(flow.TaskExecuteResumeEvent).
		Do(hook.WhenTaskExecuteResume)

	return m

}
