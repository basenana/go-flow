package controller

import (
	"context"
	"fmt"
	"github.com/zwwhdls/go-flow/eventbus"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
	"github.com/zwwhdls/go-flow/plugin"
	"time"
)

const DefaultTaskBatchInterval = 15 * time.Second

func (c *FlowController) NewFlow(builder plugin.FlowBuilder) flow.Flow {
	f := builder.Build()
	ctx := flow.FlowContext{
		Context: context.TODO(),
		FlowId:  f.ID(),
	}
	w := flowWarp{
		Flow:    f,
		ctx:     ctx,
		machine: initFlowFsm(ctx, f, builder),
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

	eventbus.Publish(flow.GetFlowTopic(flow.TaskTriggerEventTopicTpl, f.ID()), f)
	go c.triggerFlow(f.ctx, f)
	return nil
}

func (c *FlowController) triggerFlow(ctx flow.FlowContext, warp *flowWarp) {
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
		eventbus.Publish(flow.GetFlowTopic(flow.ExecutePauseEventTopicTpl, f.ID()), f)
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
		eventbus.Publish(flow.GetFlowTopic(flow.ExecuteCancelEventTopicTpl, f.ID()), f)
		return nil
	default:
		return fmt.Errorf("flow current is %s, can not cancel", f.GetStatus())
	}
}

func initFlowFsm(ctx flow.FlowContext, f flow.Flow, builder plugin.FlowBuilder) *fsm.FSM {
	hook := buildHookWithDefault(ctx, f, builder.GetFlowHook(f))

	m := fsm.New(fsm.Option{
		Name:   fmt.Sprintf("flow.%s", f.ID()),
		Obj:    f,
		Logger: nil,
	})

	m.From([]string{flow.CreatingStatus}).
		To(flow.InitializingStatus).
		When(flow.GetFlowTopic(flow.TriggerEventTopicTpl, f.ID())).
		Do(hook.WhenTrigger)

	m.From([]string{flow.InitializingStatus}).
		To(flow.RunningStatus).
		When(flow.GetFlowTopic(flow.TaskInitFinishEventTopicTpl, f.ID())).
		Do(hook.WhenInitFinish)

	m.From([]string{flow.RunningStatus}).
		To(flow.SucceedStatus).
		When(flow.GetFlowTopic(flow.ExecuteSucceedEventTopicTpl, f.ID())).
		Do(hook.WhenExecuteSucceed)

	m.From([]string{flow.InitializingStatus, flow.RunningStatus}).
		To(flow.FailedStatus).
		When(flow.GetFlowTopic(flow.ExecuteFailedEventTopicTpl, f.ID())).
		Do(hook.WhenExecuteFailed)

	m.From([]string{flow.CreatingStatus, flow.InitializingStatus, flow.RunningStatus}).
		To(flow.CanceledStatus).
		When(flow.GetFlowTopic(flow.ExecuteCancelEventTopicTpl, f.ID())).
		Do(hook.WhenExecuteCancel)

	m.From([]string{flow.RunningStatus}).
		To(flow.PausedStatus).
		When(flow.GetFlowTopic(flow.ExecutePauseEventTopicTpl, f.ID())).
		Do(hook.WhenExecutePause)

	m.From([]string{flow.PausedStatus}).
		To(flow.RunningStatus).
		When(flow.GetFlowTopic(flow.ExecuteResumeEventTopicTpl, f.ID())).
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

func statusChMerge(src, dst chan string) {
	go func() {
		for obj := range dst {
			src <- obj
		}
	}()
}
