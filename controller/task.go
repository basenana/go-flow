package controller

import (
	"context"
	"fmt"
	"github.com/zwwhdls/go-flow/eventbus"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
	"sync"
)

func (c *FlowController) triggerTask(ctx context.Context, f *flowWarp, t flow.Task) {
	c.Logger.Infof("Trigger task %s of flow %s", t.Name(), f.ID())
	flowId := f.ID()
	tName := t.Name()
	tCtx := flow.TaskContext{
		Mutex:    sync.Mutex{},
		Context:  ctx,
		Flow:     f.ctx,
		TaskName: tName,
		Operate:  flow.TaskInitOperate,
	}
	twarp := taskWarp{
		Task:    t,
		ctx:     &tCtx,
		fid:     flowId,
		machine: initTaskFsm(&tCtx, f, t),
	}
	t.SetStatus(flow.PendingStatus)

	taskCh := make(chan string, 1)
	f.machine.RegisterStatusCh(taskCh)
	twarp.machine.RegisterStatusCh(taskCh)
	twarp.machine.RegisterStatusCh(f.flowCh)

	go func() {
		var s string
		for {
			select {
			case <-tCtx.Context.Done():
				c.publishTask(flowId, tName, flow.GetTaskTopic(flow.TaskExecuteCancelEventTopicTpl, flowId, tName), f)
				return
			case s = <-taskCh:
				switch s {
				case flow.FailedStatus:
					if twarp.GetStatus() == flow.RunningStatus {
						c.publishTask(flowId, tName, flow.GetTaskTopic(flow.TaskExecutePauseEventTopicTpl, flowId, tName), f)
					}
					if twarp.GetStatus() == flow.TaskFailedStatus && twarp.ctx.Retry <= twarp.ctx.MaxRetry {
						func() {
							twarp.ctx.Mutex.Lock()
							defer twarp.ctx.Mutex.Unlock()
							c.publishTask(flowId, tName, flow.GetTaskTopic(flow.TaskExecuteRetryEventTopicTpl, flowId, tName), f)
							twarp.ctx.Retry++
						}()
					}
					continue
				case flow.PausedStatus:
					if twarp.GetStatus() == flow.RunningStatus {
						c.publishTask(flowId, tName, flow.GetTaskTopic(flow.TaskExecutePauseEventTopicTpl, flowId, tName), f)
					}
					continue
				case flow.RunningStatus:
					if t.GetStatus() != flow.PausedStatus && t.GetStatus() != flow.FailedStatus {
						continue
					}
					if twarp.ctx.Operate == flow.TaskInitOperate {
						c.publishTask(flowId, tName, flow.GetTaskTopic(flow.TaskReTriggerEventTopicTpl, flowId, tName), f)
					}
					if twarp.ctx.Operate == flow.TaskExecuteOperate {
						c.publishTask(flowId, tName, flow.GetTaskTopic(flow.TaskExecuteRetryEventTopicTpl, flowId, tName), f)
					}
					continue
				case flow.TaskSucceedStatus:
					c.Logger.Infof("task %s of flow %s succeed. tear down.", string(flowId), string(tName))
					t.Teardown(&tCtx)
					return
				case flow.TaskFailedStatus:
					c.Logger.Infof("task %s of flow %s failed. tear down.", string(flowId), string(tName))
					t.Teardown(&tCtx)
					return
				case flow.TaskInitializingStatus:
					err := t.Setup(&tCtx)
					func() {
						defer twarp.ctx.Mutex.Unlock()
						twarp.ctx.Mutex.Lock()
						twarp.ctx.Operate = flow.TaskInitOperate
					}()
					if err != nil {
						c.Logger.Errorf("Task %s of %s init error: %v", tName, flowId, err)
						c.publishTask(flowId, tName, flow.GetTaskTopic(flow.TaskExecuteFailedEventTopicTpl, flowId, tName), f)
					} else {
						c.publishTask(flowId, tName, flow.GetTaskTopic(flow.TaskInitFinishEventTopicTpl, flowId, tName), f)
					}
				case flow.TaskRunningStatus:
					c.Logger.Infof("task %s of flow %s execute", tName, flowId)
					func() {
						defer twarp.ctx.Mutex.Unlock()
						twarp.ctx.Mutex.Lock()
						twarp.ctx.Operate = flow.TaskExecuteOperate
					}()
					go twarp.Do(twarp.ctx)
					continue
				}
			}
		}
	}()

	c.publishTask(flowId, tName, flow.GetTaskTopic(flow.TaskTriggerEventTopicTpl, flowId, tName), f)
}

func (c *FlowController) publishTask(fId flow.FID, tName flow.TName, t eventbus.Topic, args ...interface{}) {
	c.Logger.Infof("Publish task [%s] event of task %s, flow %s", t, tName, fId)
	eventbus.Publish(t, args)
}

func initTaskFsm(ctx *flow.TaskContext, f *flowWarp, t flow.Task) *fsm.FSM {
	hook := buildTaskHookWithDefault(ctx, t, f.builder.GetTaskHook(f.Flow, t))
	m := fsm.New(fsm.Option{
		Name:   fmt.Sprintf("task.%s.%s", f.ID(), t.Name()),
		Obj:    t,
		Logger: nil,
	})

	m.From([]string{flow.PendingStatus}).
		To(flow.TaskInitializingStatus).
		When(flow.GetTaskTopic(flow.TaskTriggerEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskTrigger)

	m.From([]string{flow.TaskInitializingStatus}).
		To(flow.TaskRunningStatus).
		When(flow.GetTaskTopic(flow.TaskInitFinishEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskInitFinish)

	m.From([]string{flow.TaskRunningStatus}).
		To(flow.TaskSucceedStatus).
		When(flow.GetTaskTopic(flow.TaskExecuteSucceedEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskExecuteSucceed)

	m.From([]string{flow.TaskInitializingStatus, flow.TaskRunningStatus}).
		To(flow.TaskFailedStatus).
		When(flow.GetTaskTopic(flow.TaskExecuteFailedEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenExecuteFailed)

	m.From([]string{flow.TaskFailedStatus}).
		To(flow.TaskInitializingStatus).
		When(flow.GetTaskTopic(flow.TaskReTriggerEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskExecuteResume)

	m.From([]string{flow.TaskFailedStatus}).
		To(flow.TaskRunningStatus).
		When(flow.GetTaskTopic(flow.TaskExecuteRetryEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskExecuteResume)

	m.From([]string{flow.PausedStatus, flow.TaskInitializingStatus, flow.TaskRunningStatus}).
		To(flow.CanceledStatus).
		When(flow.GetTaskTopic(flow.TaskExecuteCancelEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskExecuteCancel)

	m.From([]string{flow.TaskRunningStatus}).
		To(flow.PausedStatus).
		When(flow.GetTaskTopic(flow.TaskExecutePauseEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskExecutePause)

	m.From([]string{flow.PausedStatus}).
		To(flow.TaskRunningStatus).
		When(flow.GetTaskTopic(flow.TaskExecuteResumeEventTopicTpl, f.ID(), t.Name())).
		Do(hook.WhenTaskExecuteResume)
	return m
}
