package controller

import (
	"context"
	"fmt"
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
	ctx := flow.TaskContext{}
	eventbus.Publish(flow.GetTaskTopic(flow.TaskTriggerEventTopicTpl, f.ID(), t.Name()), f)
	twarp.Do(ctx)
}

func initTaskFsm(f *flowWarp, t flow.Task) *fsm.FSM {
	ctx := flow.TaskContext{
		Context:  context.TODO(),
		Flow:     f.ctx,
		TaskName: t.Name(),
	}
	hook := buildTaskHookWithDefault(ctx, t, f.builder.GetTaskHook(f.Flow, t))
	m := fsm.New(fsm.Option{
		Name:   fmt.Sprintf("task.%s.%s", f.ID(), t.Name()),
		Obj:    f,
		Logger: nil,
	})

	m.From([]string{flow.PendingStatus}).
		To(flow.InitializingStatus).
		When(flow.GetTaskTopic(flow.TaskTriggerEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskTrigger)

	m.From([]string{flow.InitializingStatus}).
		To(flow.RunningStatus).
		When(flow.GetTaskTopic(flow.TaskInitFinishEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskInitFinish)

	m.From([]string{flow.RunningStatus}).
		To(flow.SucceedStatus).
		When(flow.GetTaskTopic(flow.TaskExecuteSucceedEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskExecuteSucceed)

	m.From([]string{flow.InitializingStatus, flow.RunningStatus}).
		To(flow.FailedStatus).
		When(flow.GetTaskTopic(flow.TaskExecuteFailedEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenExecuteFailed)

	m.From([]string{flow.PausedStatus, flow.InitializingStatus, flow.RunningStatus}).
		To(flow.CanceledStatus).
		When(flow.GetTaskTopic(flow.TaskExecuteCancelEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskExecuteCancel)

	m.From([]string{flow.RunningStatus}).
		To(flow.PausedStatus).
		When(flow.GetTaskTopic(flow.TaskExecutePauseEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskExecutePause)

	m.From([]string{flow.PausedStatus}).
		To(flow.RunningStatus).
		When(flow.GetTaskTopic(flow.TaskExecuteResumeEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskExecuteResume)

	return m

}
