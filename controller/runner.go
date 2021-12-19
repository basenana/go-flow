package controller

import (
	"context"
	"fmt"
	"github.com/zwwhdls/go-flow/eventbus"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
	"github.com/zwwhdls/go-flow/log"
	"reflect"
	"sync"
)

type runner struct {
	flow.Flow

	ctx    *flow.Context
	stopCh chan struct{}
	pause  sync.Mutex
	fsm    *fsm.FSM

	batch     []flow.Task
	batchCtx  *flow.Context
	batchCanF context.CancelFunc

	topic  eventbus.Topic
	logger log.Logger
}

func (r *runner) start(ctx *flow.Context) error {
	r.ctx = ctx
	r.topic = flow.EventTopic(r.ID())
	r.SetStatus(flow.InitializingStatus)
	r.fsm = buildFlowFSM(r)

	if err := r.Setup(ctx); err != nil {
		r.logger.Errorf("flow setup failed: %s", err.Error())
		eventbus.Publish(r.topic, fsm.Event{Type: flow.ExecuteErrorEvent, Status: r.GetStatus(), Obj: r})
		return err
	}

	r.logger.Info("flow ready to run")
	eventbus.Publish(r.topic, fsm.Event{Type: flow.TriggerEvent, Status: r.GetStatus(), Obj: r})
	return nil
}

func (r *runner) flowRun(event fsm.Event) error {
	hooks := r.GetHooks()
	if hook, ok := hooks[flow.WhenTrigger]; ok {
		if err := hook(r.ctx, r.Flow, nil); err != nil {
			r.logger.Errorf("run flow trigger hook error: %s", err.Error())
			return err
		}
	}

	if err := r.save(); err != nil {
		r.logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}

	go func() {
		r.logger.Info("flow start")
		var err error
		for {
			select {
			case <-r.ctx.Done():
				r.logger.Errorf("flow timeout")
				eventbus.Publish(r.topic, fsm.Event{Type: flow.ExecuteErrorEvent, Status: r.GetStatus(), Message: "timeout", Obj: r})
				return
			case <-r.stopCh:
				r.logger.Warn("flow closed")
				return
			default:
				r.logger.Info("run next batch")
			}

			r.pause.Lock()
			r.logger.Info("current not paused")
			needRetry := false
			for _, task := range r.batch {
				if task.GetStatus() == flow.SucceedStatus {
					continue
				}
				needRetry = true
			}

			if !needRetry {
				r.batch, err = r.NextBatch(r.ctx)
				if err != nil {
					r.logger.Errorf("got next batch error: %s", err.Error())
					eventbus.Publish(r.topic, fsm.Event{Type: flow.ExecuteErrorEvent, Status: r.GetStatus(), Message: err.Error(), Obj: r})
					return
				}

				if len(r.batch) == 0 {
					r.logger.Info("got empty batch, close flow")
					eventbus.Publish(r.topic, fsm.Event{Type: flow.ExecuteFinishEvent, Status: r.GetStatus(), Obj: r})
					break
				}
				r.logger.Infof("got new batch, contain %d tasks", len(r.batch))
			} else {
				r.logger.Warn("retry current batch")
			}
			r.pause.Unlock()

			batchCtx, canF := context.WithCancel(r.ctx.Context)
			r.batchCtx = &flow.Context{
				Context:  batchCtx,
				Logger:   r.logger,
				FlowId:   r.ID(),
				MaxRetry: r.ctx.MaxRetry,
			}
			r.batchCanF = canF

			if err = r.runBatchTasks(); err != nil {
				r.logger.Warnf("run batch failed: %s", err.Error())
			}
		}
		r.logger.Info("flow finish")
	}()

	return nil
}

func (r *runner) flowPause(event fsm.Event) error {
	r.logger.Info("flow pause")
	r.pause.Lock()
	for _, t := range r.batch {
		if t.GetStatus() == flow.RunningStatus {
			eventbus.Publish(r.topic, fsm.Event{Type: flow.TaskExecutePauseEvent, Status: t.GetStatus(), Obj: t})
		}
	}

	if err := r.save(); err != nil {
		r.logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}

	hooks := r.GetHooks()
	if hook, ok := hooks[flow.WhenExecutePause]; ok {
		if err := hook(r.ctx, r.Flow, nil); err != nil {
			r.logger.Errorf("run flow pause hook error: %s", err.Error())
			return err
		}
	}
	return nil
}

func (r *runner) flowResume(event fsm.Event) error {
	r.logger.Info("flow resume")
	for _, t := range r.batch {
		if t.GetStatus() == flow.PausedStatus {
			eventbus.Publish(r.topic, fsm.Event{Type: flow.TaskExecuteResumeEvent, Status: t.GetStatus(), Obj: t})
		}
	}
	r.pause.Unlock()

	if err := r.save(); err != nil {
		r.logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}

	hooks := r.GetHooks()
	if hook, ok := hooks[flow.WhenExecuteResume]; ok {
		if err := hook(r.ctx, r.Flow, nil); err != nil {
			r.logger.Errorf("run flow resume hook error: %s", err.Error())
			return err
		}
	}
	return nil
}

func (r *runner) flowSucceed(event fsm.Event) error {
	r.logger.Info("flow succeed")
	r.stop()

	if err := r.save(); err != nil {
		r.logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}

	hooks := r.GetHooks()
	if hook, ok := hooks[flow.WhenExecuteSucceed]; ok {
		if err := hook(r.ctx, r.Flow, nil); err != nil {
			r.logger.Errorf("run flow succeed hook error: %s", err.Error())
			return err
		}
	}
	return nil
}

func (r *runner) flowFailed(event fsm.Event) error {
	r.logger.Info("flow failed")
	r.stop()

	if err := r.save(); err != nil {
		r.logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}

	hooks := r.GetHooks()
	if hook, ok := hooks[flow.WhenExecuteFailed]; ok {
		if err := hook(r.ctx, r.Flow, nil); err != nil {
			r.logger.Errorf("run flow failed hook error: %s", err.Error())
			return err
		}
	}
	return nil
}

func (r *runner) flowCancel(event fsm.Event) error {
	r.logger.Info("flow cancel")
	if r.batchCanF != nil {
		r.batchCanF()
	}
	for _, t := range r.batch {
		switch t.GetStatus() {
		case flow.RunningStatus, flow.PausedStatus:
			t.SetStatus(flow.CanceledStatus)
		}
	}
	r.stop()

	if err := r.save(); err != nil {
		r.logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}

	hooks := r.GetHooks()
	if hook, ok := hooks[flow.WhenExecuteCancel]; ok {
		if err := hook(r.ctx, r.Flow, nil); err != nil {
			r.logger.Errorf("run flow cancel hook error: %s", err.Error())
			return err
		}
	}
	return nil
}

func (r *runner) runBatchTasks() error {

	var (
		taskFSMs    = make(map[flow.TName]*fsm.FSM)
		taskEventCh = make(chan fsm.Event, len(r.batch))
		wg          = sync.WaitGroup{}
		taskErrors  []error
	)
	defer close(taskEventCh)

	triggerTask := func(task flow.Task) {
		r.logger.Debugf("task %s build fsm", task.Name())
		tFsm := buildFlowTaskFSM(r, task)
		tCh := tFsm.EventCh()
		wg.Add(1)
		go func() {
			for evt := range tCh {
				if evt.Type == flow.TaskTriggerEvent {
					continue
				}
				taskEventCh <- evt
				r.logger.Debugf("task %s watched event: %s, and merged to task event ch", task.Name(), evt.Type)
			}
			r.logger.Debugf("task %s event watch done", task.Name())
			wg.Done()
		}()
		taskFSMs[task.Name()] = tFsm
		eventbus.Publish(r.topic, fsm.Event{Type: flow.TaskTriggerEvent, Status: task.GetStatus(), Obj: task})
	}

	for i, task := range r.batch {
		// batch retry skip already succeed task
		if task.GetStatus() == flow.SucceedStatus {
			continue
		}

		task.SetStatus(flow.InitializingStatus)
		// TODO err handler
		if err := task.Setup(r.ctx); err != nil {
			return err
		}

		r.logger.Infof("trigger task %s", task.Name())
		triggerTask(r.batch[i])
	}

	// wait all task finish
	r.logger.Infof("start waiting all task finish")
	for {
		evt := <-taskEventCh
		switch evt.Type {
		case flow.TaskExecutePauseEvent, flow.TaskExecuteResumeEvent:
			continue
		default:
		}

		eventTask := evt.Obj.(flow.Task)
		r.logger.Debugf("watched task %s event: %s", eventTask.Name(), evt.Type)

		tFsm := taskFSMs[eventTask.Name()]
		if tFsm != nil {
			_ = tFsm.Close()
			delete(taskFSMs, eventTask.Name())
		}

		if eventTask.GetStatus() == flow.FailedStatus {
			r.logger.Warnf("task %s failed: %s", eventTask.Name(), eventTask.GetMessage())
			taskErrors = append(taskErrors, fmt.Errorf("task %s failed: %s", eventTask.Name(), eventTask.GetMessage()))
		}

		if len(taskFSMs) == 0 {
			r.logger.Info("all task finish")
			break
		}
	}
	wg.Wait()

	if len(taskErrors) == 0 {
		r.batch = nil
		return nil
	}

	return fmt.Errorf("%s, and %d others", taskErrors[0], len(taskErrors))
}

func (r *runner) taskRun(event fsm.Event) error {
	task := event.Obj.(flow.Task)

	if reflect.ValueOf(task).Kind() != reflect.Ptr {
		eventbus.Publish(r.topic, fsm.Event{Type: flow.TaskExecuteErrorEvent, Status: task.GetStatus(), Message: "task obj not ptr", Obj: task})
		return fmt.Errorf("task %s obj not ptr", task.Name())
	}

	hooks := r.GetHooks()
	if hook, ok := hooks[flow.WhenTaskTrigger]; ok {
		if err := hook(r.ctx, r.Flow, task); err != nil {
			r.logger.Errorf("run task trigger hook error: %s", err.Error())
			eventbus.Publish(r.topic, fsm.Event{Type: flow.TaskExecuteErrorEvent, Status: task.GetStatus(), Message: err.Error(), Obj: task})
			return err
		}
	}

	go func() {
		var (
			currentTryTimes  = 0
			defaultRetryTime = 3
			err              error
		)
		ctx := &flow.Context{
			Context:  r.batchCtx.Context,
			FlowId:   r.batchCtx.FlowId,
			MaxRetry: defaultRetryTime,
		}
		r.logger.Infof("task %s started", task.Name())
		for {
			currentTryTimes += 1
			err = task.Do(ctx)
			if err == nil {
				r.logger.Infof("task %s succeed", task.Name())
				break
			}
			r.logger.Warnf("do task %s failed: %s", task.Name(), err.Error())
			if currentTryTimes >= ctx.MaxRetry {
				r.logger.Infof("task %s can not retry", task.Name())
				break
			}
		}

		if tStatus := waitTaskRunningOrClose(r.topic, task); tStatus != flow.RunningStatus {
			r.logger.Warnf("task was %s, stop", tStatus)
			return
		}

		if ctx.IsSucceed {
			eventbus.Publish(r.topic, fsm.Event{Type: flow.TaskExecuteErrorEvent, Status: task.GetStatus(), Message: ctx.Message, Obj: task})
			return
		}

		r.logger.Infof("run task %s finish", task.Name())
		eventbus.Publish(r.topic, fsm.Event{Type: flow.TaskExecuteFinishEvent, Status: task.GetStatus(), Obj: task})
	}()
	return nil
}

func (r *runner) taskSucceed(event fsm.Event) error {
	task := event.Obj.(flow.Task)
	r.logger.Infof("task %s succeed", task.Name())
	task.Teardown(r.batchCtx)

	if err := r.save(); err != nil {
		r.logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}

	hooks := r.GetHooks()
	if hook, ok := hooks[flow.WhenTaskExecuteSucceed]; ok {
		if err := hook(r.ctx, r.Flow, task); err != nil {
			r.logger.Errorf("run task succeed hook error: %s", err.Error())
			return err
		}
	}
	return nil
}

func (r *runner) taskFailed(event fsm.Event) error {
	task := event.Obj.(flow.Task)
	r.logger.Infof("task %s failed: %s", task.Name(), event.Message)

	msg := fmt.Sprintf("task %s failed: %s", task.Name(), event.Message)
	policy := r.GetControlPolicy()
	switch policy.FailedPolicy {
	case flow.PolicyFastFailed:
		eventbus.Publish(r.topic, fsm.Event{Type: flow.ExecuteErrorEvent, Status: r.GetStatus(), Message: msg, Obj: r})
	case flow.PolicyPaused:
		eventbus.Publish(r.topic, fsm.Event{Type: flow.ExecutePauseEvent, Status: r.GetStatus(), Message: msg, Obj: r})
	default:
		eventbus.Publish(r.topic, fsm.Event{Type: flow.ExecutePauseEvent, Status: r.GetStatus(), Message: msg, Obj: r})
	}

	task.Teardown(r.batchCtx)

	if err := r.save(); err != nil {
		r.logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}

	hooks := r.GetHooks()
	if hook, ok := hooks[flow.WhenTaskExecuteFailed]; ok {
		if err := hook(r.ctx, r.Flow, task); err != nil {
			r.logger.Errorf("run task failed hook error: %s", err.Error())
			return err
		}
	}
	return nil
}

func (r *runner) taskCanceled(event fsm.Event) error {
	task := event.Obj.(flow.Task)
	task.Teardown(r.batchCtx)
	r.logger.Infof("task %s canceled", task.Name())

	if err := r.save(); err != nil {
		r.logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}

	hooks := r.GetHooks()
	if hook, ok := hooks[flow.WhenTaskExecuteCancel]; ok {
		if err := hook(r.ctx, r.Flow, task); err != nil {
			r.logger.Errorf("run task cancel hook error: %s", err.Error())
			return err
		}
	}
	return nil
}

func (r *runner) taskPaused(event fsm.Event) error {
	task := event.Obj.(flow.Task)
	r.logger.Infof("task %s paused", task.Name())

	if err := r.save(); err != nil {
		r.logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}

	hooks := r.GetHooks()
	if hook, ok := hooks[flow.WhenTaskExecutePause]; ok {
		if err := hook(r.ctx, r.Flow, task); err != nil {
			r.logger.Errorf("run task cancel hook error: %s", err.Error())
			return err
		}
	}
	return nil
}

func (r *runner) taskResume(event fsm.Event) error {
	task := event.Obj.(flow.Task)
	r.logger.Infof("task %s resume", task.Name())

	if err := r.save(); err != nil {
		r.logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}

	hooks := r.GetHooks()
	if hook, ok := hooks[flow.WhenTaskExecuteResume]; ok {
		if err := hook(r.ctx, r.Flow, task); err != nil {
			r.logger.Errorf("run task resume hook error: %s", err.Error())
			return err
		}
	}
	return nil
}
func (r *runner) stop() {
	r.Teardown(r.ctx)
	_ = r.fsm.Close()
	close(r.stopCh)
}

func (r *runner) save() error {
	return nil
}

func buildFlowFSM(r *runner) *fsm.FSM {
	m := fsm.New(fsm.Option{
		Name: fmt.Sprintf("flow.%s", r.ID()),
		Obj:  r.Flow,
		Filter: func(event fsm.Event) bool {
			f, ok := event.Obj.(flow.Flow)
			if ok && f.ID() == r.ID() {
				return true
			}
			return false
		},
		Topic:  flow.EventTopic(r.ID()),
		Logger: r.logger.With("fsm"),
	})

	m.From([]fsm.Status{flow.InitializingStatus}).
		To(flow.RunningStatus).
		When(flow.TriggerEvent).
		Do(r.flowRun)

	m.From([]fsm.Status{flow.RunningStatus}).
		To(flow.SucceedStatus).
		When(flow.ExecuteFinishEvent).
		Do(r.flowSucceed)

	m.From([]fsm.Status{flow.InitializingStatus, flow.RunningStatus}).
		To(flow.FailedStatus).
		When(flow.ExecuteErrorEvent).
		Do(r.flowFailed)

	m.From([]fsm.Status{flow.InitializingStatus, flow.PausedStatus}).
		To(flow.CanceledStatus).
		When(flow.ExecuteCancelEvent).
		Do(r.flowCancel)

	m.From([]fsm.Status{flow.RunningStatus}).
		To(flow.PausedStatus).
		When(flow.ExecutePauseEvent).
		Do(r.flowPause)

	m.From([]fsm.Status{flow.PausedStatus}).
		To(flow.RunningStatus).
		When(flow.ExecuteResumeEvent).
		Do(r.flowResume)

	return m
}

func buildFlowTaskFSM(r *runner, t flow.Task) *fsm.FSM {
	m := fsm.New(fsm.Option{
		Name: fmt.Sprintf("flow.%s.%s", r.ID(), t.Name()),
		Obj:  t,
		Filter: func(event fsm.Event) bool {
			tObj, ok := event.Obj.(flow.Task)
			if ok && tObj.Name() == t.Name() {
				return true
			}
			return false
		},
		Topic:  flow.EventTopic(r.ID()),
		Logger: r.logger.With(fmt.Sprintf("task.%s.fsm", t.Name())),
	})

	m.From([]fsm.Status{flow.InitializingStatus}).
		To(flow.RunningStatus).
		When(flow.TaskTriggerEvent).
		Do(r.taskRun)

	m.From([]fsm.Status{flow.RunningStatus}).
		To(flow.SucceedStatus).
		When(flow.TaskExecuteFinishEvent).
		Do(r.taskSucceed)

	m.From([]fsm.Status{flow.RunningStatus}).
		To(flow.PausedStatus).
		When(flow.TaskExecutePauseEvent).
		Do(r.taskPaused)

	m.From([]fsm.Status{flow.PausedStatus}).
		To(flow.RunningStatus).
		When(flow.TaskExecuteResumeEvent).
		Do(r.taskResume)

	m.From([]fsm.Status{flow.RunningStatus, flow.PausedStatus}).
		To(flow.CanceledStatus).
		When(flow.TaskExecuteCancelEvent).
		Do(r.taskCanceled)

	m.From([]fsm.Status{flow.InitializingStatus, flow.RunningStatus}).
		To(flow.FailedStatus).
		When(flow.TaskExecuteErrorEvent).
		Do(r.taskFailed)

	return m
}
