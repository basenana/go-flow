package controller

import (
	"context"
	"fmt"
	"github.com/zwwhdls/go-flow/eventbus"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
	"github.com/zwwhdls/go-flow/plugin"
	"sync"
	"time"
)

const DefaultTaskBatchInterval = 15 * time.Second

func (c *FlowController) NewFlow(ctx context.Context, builder plugin.FlowBuilder) flow.Flow {
	f := builder.Build()
	c.Logger.Infof("Build flow %s", f.ID())
	flowCtx := flow.FlowContext{
		Mutex:   sync.Mutex{},
		Context: ctx,
		FlowId:  f.ID(),
		Operate: flow.FlowInitOperate,
	}
	w := flowWarp{
		Flow:    f,
		ctx:     &flowCtx,
		machine: initFlowFsm(&flowCtx, f, builder),
		builder: builder,
	}
	f.SetStatus(flow.CreatingStatus)
	c.flows[f.ID()] = &w
	return w
}

func (c *FlowController) TriggerFlow(flowId flow.FID) error {
	c.Logger.Infof("Trigger flow %s", flowId)
	f, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}

	go c.triggerFlow(f.ctx, f)
	return nil
}

func (c *FlowController) triggerFlow(ctx *flow.FlowContext, warp *flowWarp) {
	flowId := warp.ID()
	var (
		tasks []flow.Task
		err   error
		canf  context.CancelFunc
		timer = time.NewTimer(DefaultTaskBatchInterval)
	)
	defer timer.Stop()

	ctx.Context, canf = context.WithCancel(ctx.Context)
	warp.flowCh = make(chan string, 1)
	warp.machine.RegisterStatusCh(warp.flowCh)

	go func() {
		var s string
		var cancel context.CancelFunc
		defer canf()
		for {
			select {
			case s = <-warp.flowCh:
				c.Logger.Infof("Receive event: %v", s)
				switch s {
				case flow.PausedStatus:
					continue
				case flow.FailedStatus:
					c.Logger.Infof("flow %s failed. tear down.", string(flowId))
					warp.Teardown(ctx)
					continue
				case flow.CanceledStatus:
					cancel()
					return
				case flow.SucceedStatus:
					c.Logger.Infof("flow %s succeed. tear down.", string(flowId))
					warp.Teardown(ctx)
					return
				case flow.InitializingStatus:
					err := warp.Setup(ctx)
					func() {
						defer warp.ctx.Mutex.Unlock()
						warp.ctx.Mutex.Lock()
						warp.ctx.Operate = flow.FlowInitOperate
					}()
					if err != nil {
						c.Logger.Errorf("Flow %s init error: %v", flowId, err)
						c.publishFlow(flowId, flow.GetFlowTopic(flow.ExecuteFailedEventTopicTpl, flowId), warp)
					} else {
						c.publishFlow(flowId, flow.GetFlowTopic(flow.InitFinishEventTopicTpl, flowId), warp)
					}
					continue
				case flow.TaskSucceedStatus:
					fallthrough
				case flow.RunningStatus:
					func() {
						defer warp.ctx.Mutex.Unlock()
						warp.ctx.Mutex.Lock()
						warp.ctx.Operate = flow.FlowExecuteOperate
					}()
					if !isCurrentTaskAllFinish(warp.currentTasks) {
						c.Logger.Infof("Task not all finish.")
						continue
					}

					c.Logger.Infof("Get next batch tasks.")
					tasks, err = warp.NextBatch(ctx)
					if err != nil {
						c.Logger.Errorf("Get next batch error: %v", err)
						return
					}
					warp.currentTasks = tasks

					if len(tasks) == 0 {
						c.Logger.Infof("There are no tasks to be execute.")
						c.publishFlow(flowId, flow.GetFlowTopic(flow.ExecuteSucceedEventTopicTpl, flowId), warp)
						continue
					}

					tCtx, canc := context.WithCancel(warp.ctx.Context)
					cancel = canc
					for i := range tasks {
						task := tasks[i]
						go c.triggerTask(tCtx, warp, task)
					}
				case flow.TaskFailedStatus:
					c.publishFlow(flowId, flow.GetFlowTopic(flow.ExecuteFailedEventTopicTpl, flowId), warp)
				}
			}
		}
	}()

	c.publishFlow(flowId, flow.GetFlowTopic(flow.TriggerEventTopicTpl, flowId), warp)
}

func (c *FlowController) publishFlow(fId flow.FID, t eventbus.Topic, args ...interface{}) {
	c.Logger.Infof("Publish flow [%s] event of flow %s", t, fId)
	eventbus.Publish(t, args)
}

func (c *FlowController) PauseFlow(flowId flow.FID) error {
	c.Logger.Infof("Pause flow %s", flowId)
	f, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	if f.GetStatus() == flow.RunningStatus {
		c.publishFlow(flowId, flow.GetFlowTopic(flow.ExecutePauseEventTopicTpl, flowId), f)
	}
	return fmt.Errorf("flow current is %s, can not pause", f.GetStatus())
}

func (c *FlowController) CancelFlow(flowId flow.FID) error {
	c.Logger.Infof("Cancel flow %s", flowId)
	f, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	switch f.GetStatus() {
	case flow.CreatingStatus, flow.InitializingStatus, flow.RunningStatus:
		c.publishFlow(flowId, flow.GetFlowTopic(flow.ExecuteCancelEventTopicTpl, flowId), f)
		for i := range f.currentTasks {
			task := f.currentTasks[i]
			eventbus.Publish(flow.GetTaskTopic(flow.TaskExecuteCancelEventTopicTpl, f.ID(), task.Name()), f, task)
		}
		return nil
	default:
		return fmt.Errorf("flow current is %s, can not cancel", f.GetStatus())
	}
}

func (c *FlowController) ResumeFlow(flowId flow.FID) error {
	c.Logger.Infof("Resume flow %s", flowId)
	f, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	switch f.GetStatus() {
	case flow.PausedStatus:
		c.publishFlow(flowId, flow.GetFlowTopic(flow.ExecuteResumeEventTopicTpl, flowId), f)
		return nil
	default:
		return fmt.Errorf("flow current is %s, can not resume", f.GetStatus())
	}
}

func (c *FlowController) RetriggerFlow(flowId flow.FID) error {
	c.Logger.Infof("Retrigger flow %s", flowId)
	f, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	if f.GetStatus() == flow.FailedStatus && f.ctx.Operate == flow.FlowInitOperate {
		c.publishFlow(flowId, flow.GetFlowTopic(flow.ReTriggerEventTopicTpl, flowId), f)
		return nil
	}
	return fmt.Errorf("flow current is %s, can not retrigger", f.GetStatus())
}

func (c *FlowController) RetryFlow(flowId flow.FID) error {
	c.Logger.Infof("Retry flow %s", flowId)
	f, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	if f.GetStatus() == flow.FailedStatus && f.ctx.Operate == flow.FlowExecuteOperate {
		c.publishFlow(flowId, flow.GetFlowTopic(flow.ExecuteRetryEventTopicTpl, flowId), f)
		return nil
	}
	return fmt.Errorf("flow current is %s, can not retry", f.GetStatus())
}

func initFlowFsm(ctx *flow.FlowContext, f flow.Flow, builder plugin.FlowBuilder) *fsm.FSM {
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
		When(flow.GetFlowTopic(flow.InitFinishEventTopicTpl, f.ID())).
		Do(hook.WhenInitFinish)

	m.From([]string{flow.RunningStatus}).
		To(flow.SucceedStatus).
		When(flow.GetFlowTopic(flow.ExecuteSucceedEventTopicTpl, f.ID())).
		Do(hook.WhenExecuteSucceed)

	m.From([]string{flow.InitializingStatus, flow.RunningStatus}).
		To(flow.FailedStatus).
		When(flow.GetFlowTopic(flow.ExecuteFailedEventTopicTpl, f.ID())).
		Do(hook.WhenExecuteFailed)

	m.From([]string{flow.FailedStatus}).
		To(flow.InitializingStatus).
		When(flow.GetFlowTopic(flow.ReTriggerEventTopicTpl, f.ID())).
		Do(hook.WhenExecuteResume)

	m.From([]string{flow.FailedStatus}).
		To(flow.RunningStatus).
		When(flow.GetFlowTopic(flow.ExecuteRetryEventTopicTpl, f.ID())).
		Do(hook.WhenExecuteResume)

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
		case flow.TaskSucceedStatus, flow.TaskFailedStatus, flow.CanceledStatus:
			continue
		default:
			return false
		}
	}
	return true
}

func statusChCopy(src, dst chan string) {
	go func() {
		for obj := range dst {
			src <- obj
		}
	}()
}
