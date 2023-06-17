/*
   Copyright 2022 Go-Flow Authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package flow

import (
	"context"
	"fmt"
	"github.com/basenana/go-flow/fsm"
	"github.com/basenana/go-flow/utils"
	"reflect"
	"sync"
	"time"
)

type Runner interface {
	Start(ctx context.Context) error
	Pause() error
	Resume() error
	Cancel() error
}

func NewRunner(f *Flow, s Interface) Runner {
	return nil
}

type runner struct {
	*Flow

	ctx      context.Context
	fsm      *fsm.FSM
	dag      *DAG
	executor Executor

	batch     []runningTask
	batchCtx  context.Context
	batchCanF context.CancelFunc

	storage Interface
}

func (r *runner) Start(ctx context.Context) error {
	r.ctx = ctx
	if r.Status != InitializingStatus {
		r.SetStatus(InitializingStatus)
		if err := r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
			logger.Errorf("save flow status failed: %s")
			return err
		}
	}

	var err error
	r.dag, err = buildDAG(r.Tasks)
	if err != nil {
		r.SetStatus(FailedStatus)
		logger.Errorf("build dag failed: %s, save flow to failed: %s", err, r.storage.SaveFlow(r.ctx, r.Flow))
		return err
	}
	r.executor = NewLocalExecutor(r.Flow)
	r.fsm = buildFlowFSM(r)

	logger.Info("flow ready to run")
	if err = r.pushEvent2FlowFSM(fsm.Event{Type: TriggerEvent, Status: r.GetStatus(), Obj: r}); err != nil {
		return err
	}

	return nil
}

func (r *runner) Pause() error {
	if r.Status == RunningStatus {
		return r.pushEvent2FlowFSM(fsm.Event{Type: ExecutePauseEvent, Obj: r.Flow})
	}
	return nil
}

func (r *runner) Resume() error {
	if r.Status == PausedStatus {
		return r.pushEvent2FlowFSM(fsm.Event{Type: ExecuteResumeEvent, Obj: r.Flow})
	}
	return nil
}

func (r *runner) Cancel() error {
	if r.Status == SucceedStatus || r.Status == FailedStatus || r.Status == ErrorStatus {
		return nil
	}
	return r.pushEvent2FlowFSM(fsm.Event{Type: ExecuteCancelEvent, Obj: r.Flow})
}

func (r *runner) flowRun() error {
	go func() {
		logger.Info("flow start")

		var err error
		for {
			select {
			case <-r.ctx.Done():
				logger.Errorf("flow timeout")
				_ = r.pushEvent2FlowFSM(fsm.Event{Type: ExecuteErrorEvent, Status: r.GetStatus(), Message: "timeout", Obj: r})
				return
			case <-r.ctx.Done():
				logger.Warn("flow closed")
				return
			default:
				logger.Info("run next batch")
			}

		StatusCheckLoop:
			for {
				switch r.GetStatus() {
				case RunningStatus:
					break StatusCheckLoop
				case FailedStatus:
					logger.Warn("flow closed")
					return
				}
				time.Sleep(15 * time.Second)
			}

			logger.Info("current not paused")
			needRetry := false
			for _, task := range r.batch {
				if task.GetStatus() == SucceedStatus {
					continue
				}
				needRetry = true
			}

			if !needRetry {
				if err = r.makeNextBatch(); err != nil {
					logger.Errorf("make next batch plan error: %s, stop flow.", err.Error())
					_ = r.pushEvent2FlowFSM(fsm.Event{Type: ExecuteErrorEvent, Status: r.GetStatus(), Message: err.Error(), Obj: r})
					return
				}
				if len(r.batch) == 0 {
					logger.Info("got empty batch, close finished flow")
					_ = r.pushEvent2FlowFSM(fsm.Event{Type: ExecuteFinishEvent, Status: r.GetStatus(), Message: "finish", Obj: r})
					break
				}
			} else {
				logger.Warn("retry current batch")
			}

			batchCtx, canF := context.WithCancel(r.ctx)
			r.batchCtx = batchCtx
			r.batchCanF = canF

			if err = r.runBatchTasks(); err != nil {
				logger.Warnf("run batch failed: %s", err.Error())
				_ = r.pushEvent2FlowFSM(fsm.Event{Type: ExecuteErrorEvent, Status: r.GetStatus(), Message: err.Error(), Obj: r})
				return
			}
		}
		logger.Info("flow finish")
	}()

	return nil
}

func (r *runner) makeNextBatch() error {
	var (
		nextTasks []*Task
		err       error
	)

	for len(nextTasks) == 0 {
		nextTasks, err = r.nextBatch()
		if err != nil {
			logger.Errorf("got next batch error: %s", err.Error())
			return err
		}

		if len(nextTasks) == 0 {
			// all task finish
			return nil
		}

		newBatch := make([]runningTask, 0, len(nextTasks))
		for i, task := range nextTasks {
			if IsFinishedStatus(task.Status) {
				continue
			}

			newBatch = append(newBatch, runningTask{
				Task: nextTasks[i],
				FSM:  buildFlowTaskFSM(r, nextTasks[i]),
			})
		}
		r.batch = newBatch
	}
	logger.Infof("got new batch, contain %d tasks", len(r.batch))
	return nil
}

func (r *runner) nextBatch() ([]*Task, error) {
	taskTowards := r.dag.nextBatchTasks()

	nextBatch := make([]*Task, 0, len(taskTowards))
	for _, t := range taskTowards {
		for i := range r.Tasks {
			task := r.Tasks[i]
			if task.Name == t.taskName {
				nextBatch = append(nextBatch, &task)
				break
			}
		}
	}
	return nextBatch, nil
}

func (r *runner) handleFlowRun(event fsm.Event) error {
	return r.flowRun()
}

func (r *runner) handleFlowPause(event fsm.Event) error {
	logger.Info("flow pause")

	if err := r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
		logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}
	return nil
}

func (r *runner) handleFlowResume(event fsm.Event) error {
	logger.Info("flow resume")

	if err := r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
		logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}

	return nil
}

func (r *runner) handleFlowSucceed(event fsm.Event) error {
	logger.Info("flow succeed")
	r.stop("succeed")

	if err := r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
		logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}
	return nil
}

func (r *runner) handleFlowFailed(event fsm.Event) error {
	logger.Info("flow failed")
	r.stop(event.Message)

	if err := r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
		logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}
	return nil
}

func (r *runner) handleFlowCancel(event fsm.Event) error {
	logger.Info("flow cancel")
	if r.batchCanF != nil {
		r.batchCanF()
	}

	cancelTasksErr := utils.NewErrors()
	for _, t := range r.batch {
		switch t.GetStatus() {
		case InitializingStatus, RunningStatus, PausedStatus:
			logger.Infof("cancel %s task %s", t.GetStatus(), t.Name)
			if err := t.Event(fsm.Event{Type: TaskExecuteCancelEvent, Status: t.GetStatus(), Obj: t.Task}); err != nil {
				cancelTasksErr = append(cancelTasksErr, err)
			}
		}
	}
	if cancelTasksErr.IsError() {
		return cancelTasksErr
	}

	r.stop(event.Message)

	if err := r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
		logger.Errorf("save flow status failed: %s", err.Error())
		return err
	}

	return nil
}

func (r *runner) runBatchTasks() error {
	var (
		wg         = sync.WaitGroup{}
		taskErrors = utils.NewErrors()
	)

	triggerTask := func(task runningTask) error {
		logger.Infof("trigger task %s", task.Name)
		if IsFinishedStatus(task.GetStatus()) {
			logger.Warnf("task %s was finished(%s)", task.Name, task.GetStatus())
			return nil
		}

		select {
		case <-r.ctx.Done():
			logger.Infof("flow was closed, skip task %s trigger", task.Name)
			return nil
		default:
			if err := task.Event(fsm.Event{Type: TaskTriggerEvent, Status: task.GetStatus(), Obj: task.Task}); err != nil {
				switch r.GetStatus() {
				case CanceledStatus, FailedStatus:
					return nil
				default:
					return err
				}
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				if runTaskErr := r.taskRun(task); runTaskErr != nil {
					taskErrors = append(taskErrors, runTaskErr)
				}
			}()
			return nil
		}
	}

	for i, task := range r.batch {
		// batch retry skip already succeed task
		if task.GetStatus() == SucceedStatus {
			continue
		}

		task.SetStatus(InitializingStatus)
		if err := triggerTask(r.batch[i]); err != nil {
			return err
		}
	}

	// wait all task finish
	logger.Infof("start waiting all task finish")
	wg.Wait()

	if len(taskErrors) == 0 {
		r.batch = nil
		return nil
	}

	return fmt.Errorf("%s, and %d others", taskErrors[0], len(taskErrors))
}

func (r *runner) taskRun(task runningTask) error {
	var (
		currentTryTimes = 0
		err             error
	)
	logger.Infof("task %s started", task.Name)
	for {
		currentTryTimes += 1
		err = r.executor.DoOperation(r.batchCtx, task.OperatorSpec)
		if err == nil {
			logger.Infof("task %s succeed", task.Name)
			break
		}
		logger.Warnf("do task %s failed: %s", task.Name, err.Error())
		if currentTryTimes >= task.RetryOnFailed {
			logger.Infof("task %s can not retry", task.Name)
			break
		}
	}

	if err != nil {
		msg := fmt.Sprintf("task %s failed: %s", task.Name, err)
		policy := r.ControlPolicy

		updateStatsErrors := utils.NewErrors()
		switch policy.FailedPolicy {
		case PolicyFastFailed:
			err = r.pushEvent2FlowFSM(fsm.Event{Type: ExecuteFailedEvent, Status: r.GetStatus(), Message: msg, Obj: r})
		case PolicyPaused:
			err = r.pushEvent2FlowFSM(fsm.Event{Type: ExecutePauseEvent, Status: r.GetStatus(), Message: msg, Obj: r})
		default:
			err = r.pushEvent2FlowFSM(fsm.Event{Type: ExecutePauseEvent, Status: r.GetStatus(), Message: msg, Obj: r})
		}

		if err != nil {
			updateStatsErrors = append(updateStatsErrors, err)
		}
		if err = task.Event(fsm.Event{Type: TaskExecuteErrorEvent, Status: task.GetStatus(), Message: msg, Obj: task}); err != nil {
			updateStatsErrors = append(updateStatsErrors, err)
		}

		if updateStatsErrors.IsError() {
			return updateStatsErrors
		}
		return nil
	}

	logger.Infof("run task %s finish", task.Name)
	return task.Event(fsm.Event{Type: TaskExecuteFinishEvent, Status: task.GetStatus(), Obj: task})
}

func (r *runner) handleTaskRun(event fsm.Event) error {
	task, ok := event.Obj.(*Task)
	if !ok || reflect.ValueOf(task).Kind() != reflect.Ptr {
		return fmt.Errorf("task %s obj not ptr", task.Name)
	}

	if err := r.storage.SaveTask(r.ctx, r.ID, task); err != nil {
		logger.Errorf("save task status failed: %s", err.Error())
		return err
	}
	return nil
}

func (r *runner) handleTaskSucceed(event fsm.Event) error {
	task := event.Obj.(*Task)
	logger.Infof("task %s succeed", task.Name)

	r.dag.updateTaskStatus(task.Name, task.Status)
	if err := r.storage.SaveTask(r.ctx, r.ID, task); err != nil {
		logger.Errorf("save task status failed: %s", err.Error())
		return err
	}
	return nil
}

func (r *runner) handleTaskFailed(event fsm.Event) (err error) {
	task := event.Obj.(*Task)
	logger.Infof("task %s failed: %s", task.Name, event.Message)

	r.dag.updateTaskStatus(task.Name, task.Status)
	if err = r.storage.SaveTask(r.ctx, r.ID, task); err != nil {
		logger.Errorf("save task status failed: %s", err.Error())
		return err
	}
	return nil
}

func (r *runner) handleTaskCanceled(event fsm.Event) error {
	task := event.Obj.(*Task)
	logger.Infof("task %s canceled", task.Name)

	r.dag.updateTaskStatus(task.Name, task.Status)
	if err := r.storage.SaveTask(r.ctx, r.ID, task); err != nil {
		logger.Errorf("save task status failed: %s", err.Error())
		return err
	}
	return nil
}

func (r *runner) stop(msg string) {
	logger.Debugf("stopping flow: %s", msg)
	r.SetMessage(msg)

	func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("teardown panic: %v", err)
			}
		}()
	}()
}

func (r *runner) pushEvent2FlowFSM(event fsm.Event) error {
	err := r.fsm.Event(event)
	if err != nil {
		logger.Errorf("push event to flow FSM with event %s error: %s, msg: %s", event.Type, err.Error(), event.Message)
		return err
	}
	return nil
}

type runningTask struct {
	*Task
	*fsm.FSM
}

func buildFlowFSM(r *runner) *fsm.FSM {
	m := fsm.New(fsm.Option{
		Obj:    r.Flow,
		Logger: logger.With("fsm"),
	})

	m.From([]string{InitializingStatus}).
		To(RunningStatus).
		When(TriggerEvent).
		Do(r.handleFlowRun)

	m.From([]string{RunningStatus}).
		To(SucceedStatus).
		When(ExecuteFinishEvent).
		Do(r.handleFlowSucceed)

	m.From([]string{InitializingStatus, RunningStatus}).
		To(FailedStatus).
		When(ExecuteErrorEvent).
		Do(r.handleFlowFailed)

	m.From([]string{InitializingStatus, PausedStatus}).
		To(CanceledStatus).
		When(ExecuteCancelEvent).
		Do(r.handleFlowCancel)

	m.From([]string{RunningStatus}).
		To(PausedStatus).
		When(ExecutePauseEvent).
		Do(r.handleFlowPause)

	m.From([]string{PausedStatus}).
		To(RunningStatus).
		When(ExecuteResumeEvent).
		Do(r.handleFlowResume)

	return m
}

func buildFlowTaskFSM(r *runner, t *Task) *fsm.FSM {
	m := fsm.New(fsm.Option{
		Obj:    t,
		Logger: logger.With(fmt.Sprintf("task.%s.fsm", t.Name)),
	})

	m.From([]string{InitializingStatus}).
		To(RunningStatus).
		When(TaskTriggerEvent).
		Do(r.handleTaskRun)

	m.From([]string{RunningStatus}).
		To(SucceedStatus).
		When(TaskExecuteFinishEvent).
		Do(r.handleTaskSucceed)

	m.From([]string{InitializingStatus, RunningStatus, PausedStatus}).
		To(CanceledStatus).
		When(TaskExecuteCancelEvent).
		Do(r.handleTaskCanceled)

	m.From([]string{InitializingStatus, RunningStatus}).
		To(FailedStatus).
		When(TaskExecuteErrorEvent).
		Do(r.handleTaskFailed)

	return m
}
