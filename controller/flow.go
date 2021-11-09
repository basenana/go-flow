package controller

import (
	"context"
	"fmt"
	taskcontext "go-flow/context"
	"go-flow/eventbus"
	"go-flow/flow"
	"go-flow/fsm"
	"go-flow/plugin"
	"time"
)

const DefaultTaskBatchInterval = 15 * time.Second

func (c *FlowController) NewFlow(builder plugin.FlowBuilder) flow.Flow {
	f := builder.Build()
	w := flowWarp{
		Flow:    f,
		machine: initFlowFsm(f, builder),
		builder: builder,
	}
	f.SetStatus(flow.CreatingStatus)
	c.flows[f.ID()] = &w
	return w
}

func (c *FlowController) TriggerFlow(flowId flow.FID) error {
	f, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}

	ctx := taskcontext.FlowContext{}
	eventbus.Publish(flow.TriggerEvent, f)
	go c.triggerFlow(ctx, f)
	return nil
}

func (c *FlowController) triggerFlow(ctx taskcontext.FlowContext, warp *flowWarp) {
	var (
		tasks []flow.Task
		err   error
		canf  context.CancelFunc
		timer = time.NewTimer(DefaultTaskBatchInterval)
	)
	defer timer.Stop()

	ctx.Context, canf = context.WithCancel(ctx.Context)
	defer canf()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		switch warp.GetStatus() {
		case flow.PausedStatus:
			continue
		case flow.FailedStatus, flow.CanceledStatus:
			return
		}

		if !isCurrentTaskAllFinish(warp.currentTasks) {
			continue
		}

		tasks, err = warp.NextBatch(ctx)
		if err != nil {
			return
		}
		warp.currentTasks = tasks

		for i := range tasks {
			task := tasks[i]
			go c.triggerTask(warp, task)
		}
	}
}

func (c *FlowController) PauseFlow(flowId flow.FID) error {
	f, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	if f.GetStatus() == flow.RunningStatus {
		eventbus.Publish(flow.ExecutePauseEvent, f)
	}
	return fmt.Errorf("flow current is %s, can not pause", f.GetStatus())
}

func (c *FlowController) CancelFlow(flowId flow.FID) error {
	f, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	switch f.GetStatus() {
	case flow.CreatingStatus, flow.InitializingStatus, flow.RunningStatus:
		eventbus.Publish(flow.ExecuteCancelEvent, f)
		return nil
	default:
		return fmt.Errorf("flow current is %s, can not cancel", f.GetStatus())
	}
}

func initFlowFsm(f flow.Flow, builder plugin.FlowBuilder) *fsm.FSM {
	hook := builder.GetFlowHook(f)
	hook = buildHookWithDefault(hook)

	m := fsm.New(fsm.Option{
		Name:   fmt.Sprintf("flow.%s", f.ID()),
		Obj:    f,
		Logger: nil,
	})

	m.From([]string{flow.CreatingStatus}).
		To(flow.InitializingStatus).
		When(flow.TriggerEvent).
		Do(hook.WhenTrigger)

	m.From([]string{flow.InitializingStatus}).
		To(flow.RunningStatus).
		When(flow.InitFinishEvent).
		Do(hook.WhenInitFinish)

	m.From([]string{flow.RunningStatus}).
		To(flow.SucceedStatus).
		When(flow.ExecuteSucceedEvent).
		Do(hook.WhenExecuteSucceed)

	m.From([]string{flow.InitializingStatus, flow.RunningStatus}).
		To(flow.FailedStatus).
		When(flow.ExecuteFailedEvent).
		Do(hook.WhenExecuteFailed)

	m.From([]string{flow.CreatingStatus, flow.InitializingStatus, flow.RunningStatus}).
		To(flow.CanceledStatus).
		When(flow.ExecuteCancelEvent).
		Do(hook.WhenExecuteCancel)

	m.From([]string{flow.RunningStatus}).
		To(flow.PausedStatus).
		When(flow.ExecutePauseEvent).
		Do(hook.WhenExecutePause)

	m.From([]string{flow.PausedStatus}).
		To(flow.RunningStatus).
		When(flow.ExecuteResumeEvent).
		Do(hook.WhenExecuteResume)

	return m
}

func isCurrentTaskAllFinish(tasks []flow.Task) bool {
	for t := range tasks {
		task := tasks[t]
		s := task.GetStatus()
		switch s {
		case flow.SucceedStatus, flow.FailedStatus, flow.CanceledStatus:
			continue
		default:
			return false
		}
	}
	return true
}
