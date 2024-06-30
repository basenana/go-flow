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
	"errors"
	"fmt"
	"github.com/basenana/go-flow/fsm"
	"github.com/basenana/go-flow/types"
	"github.com/basenana/go-flow/utils"
	"sync"
	"time"
)

type Runner interface {
	Start(ctx context.Context) error
	Pause() error
	Resume() error
	Cancel() error
}

func NewRunner(f types.Flow, storage Storage) Runner {
	return &runner{
		Flow:    f,
		mux:     sync.Mutex{},
		storage: storage,
		logger:  utils.NewLogger("runner").With(f.GetID()),
	}
}

type runner struct {
	types.Flow

	ctx      context.Context
	canF     context.CancelFunc
	fsm      *fsm.FSM
	dag      *DAG
	executor Executor
	stopCh   chan struct{}
	started  bool
	mux      sync.Mutex

	storage Storage
	logger  utils.Logger
}

func (r *runner) Start(ctx context.Context) (err error) {
	if IsFinishedStatus(r.Flow.GetStatus()) {
		return
	}

	startAt := time.Now()
	r.logger.Info("start runner")
	r.ctx, r.canF = context.WithCancel(ctx)
	if err = r.initial(); err != nil {
		r.logger.Errorf("job initial failed: %s", err)
		return err
	}
	defer func() {
		if deferErr := r.storage.SaveFlow(ctx, r.Flow); deferErr != nil {
			r.logger.Errorf("save job to metabase failed: %s", deferErr)
		}
		execDu := time.Since(startAt)
		r.logger.Infof("close runner, cost %s", execDu.String())
	}()

	r.dag, err = buildDAG(r.Flow.Tasks())
	if err != nil {
		r.Flow.SetStatus(types.FailedStatus)
		r.Flow.SetMessage("build dag failed")
		r.logger.Errorf("build dag failed: %s", err)
		return err
	}
	r.executor, err = newExecutor(r.Flow.GetExecutor(), r.Flow)
	if err != nil {
		r.Flow.SetStatus(types.FailedStatus)
		r.Flow.SetMessage("build executor failed")
		r.logger.Errorf("build executor failed: %s", err)
		return err
	}
	r.fsm = buildFlowFSM(r)
	r.stopCh = make(chan struct{})

	if !r.waitingForRunning(ctx) {
		return
	}

	if err = r.executor.Setup(ctx); err != nil {
		r.Flow.SetStatus(types.ErrorStatus)
		r.Flow.SetMessage(err.Error())
		r.logger.Errorf("setup executor failed: %s", err)
		return err
	}

	if err = r.pushEvent2FlowFSM(fsm.Event{Type: types.TriggerEvent, Status: r.Flow.GetStatus(), Obj: r.Flow}); err != nil {
		r.Flow.SetStatus(types.ErrorStatus)
		r.Flow.SetMessage(err.Error())
		r.logger.Errorf("trigger job failed %s", err)
		return err
	}

	// waiting all task down
	<-r.stopCh

	r.executor.Teardown(ctx)
	return nil
}

func (r *runner) Pause() error {
	if r.Flow.GetStatus() == types.RunningStatus {
		return r.pushEvent2FlowFSM(fsm.Event{Type: types.ExecutePauseEvent, Obj: r.Flow})
	}
	return nil
}

func (r *runner) Resume() error {
	if r.Flow.GetStatus() == types.PausedStatus {
		return r.pushEvent2FlowFSM(fsm.Event{Type: types.ExecuteResumeEvent, Obj: r.Flow})
	}
	return nil
}

func (r *runner) Cancel() error {
	switch r.Flow.GetStatus() {
	case types.SucceedStatus, types.FailedStatus, types.ErrorStatus:
		return nil
	}
	return r.pushEvent2FlowFSM(fsm.Event{Type: types.ExecuteCancelEvent, Obj: r.Flow})
}

func (r *runner) handleJobRun(event fsm.Event) error {
	r.logger.Info("job ready to run")
	if err := r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
		r.logger.Errorf("save job status failed: %s", err)
		return err
	}

	r.mux.Lock()
	defer r.mux.Unlock()
	if r.started {
		return nil
	}
	r.started = true

	go func() {
		var (
			isFinish bool
			err      error
		)
		defer func() {
			r.logger.Info("job finished")
			close(r.stopCh)
		}()

		for {
			select {
			case <-r.ctx.Done():
				err = r.ctx.Err()
				r.logger.Errorf("job timeout")
			default:
				isFinish, err = r.batchRun()
			}

			if err != nil {
				if errors.Is(err, context.Canceled) {
					r.logger.Errorf("job canceled")
					return
				}
				_ = r.pushEvent2FlowFSM(fsm.Event{Type: types.ExecuteErrorEvent, Status: r.GetStatus(), Message: err.Error(), Obj: r.Flow})
				return
			}
			if isFinish {
				if !IsFinishedStatus(r.GetStatus()) {
					_ = r.pushEvent2FlowFSM(fsm.Event{Type: types.ExecuteFinishEvent, Status: r.GetStatus(), Message: "finish", Obj: r.Flow})
				}
				return
			}
		}
	}()
	return nil
}

func (r *runner) handleJobPause(event fsm.Event) error {
	r.logger.Info("job pause")

	if err := r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
		r.logger.Errorf("save job status failed: %s", err)
		return err
	}
	return nil
}

func (r *runner) handleJobResume(event fsm.Event) error {
	r.logger.Info("job resume")

	if err := r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
		r.logger.Errorf("save job status failed: %s", err)
		return err
	}

	return nil
}

func (r *runner) handleJobSucceed(event fsm.Event) error {
	r.logger.Info("job succeed")
	if err := r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
		r.logger.Errorf("save job status failed: %s", err)
		return err
	}
	r.close("succeed")
	return nil
}

func (r *runner) handleJobFailed(event fsm.Event) error {
	r.logger.Info("job failed")
	if err := r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
		r.logger.Errorf("save job status failed: %s", err)
		return err
	}
	r.close(event.Message)
	return nil
}

func (r *runner) handleJobCancel(event fsm.Event) error {
	r.logger.Info("job cancel")
	if err := r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
		r.logger.Errorf("save job status failed: %s", err)
		return err
	}
	r.close(event.Message)
	return nil
}

func (r *runner) batchRun() (finish bool, err error) {
	if !r.waitingForRunning(r.ctx) {
		return true, nil
	}

	defer func() {
		if panicErr := utils.Recover(); panicErr != nil {
			err = panicErr
		}
	}()

	var batch []types.Task
	if batch, err = r.nextBatch(); err != nil {
		r.logger.Errorf("make next batch plan error: %s, stop job.", err)
		return
	}

	if len(batch) == 0 {
		r.logger.Info("all batch finished, close job")
		return true, nil
	}

	batchCtx, batchCanF := context.WithCancel(r.ctx)
	defer batchCanF()
	wg := sync.WaitGroup{}
	for i := range batch {
		wg.Add(1)
		go func(t types.Task) {
			defer wg.Done()
			if needCancel := r.taskRun(batchCtx, t); needCancel {
				batchCanF()
			}
		}(batch[i])
	}
	wg.Wait()

	return false, nil
}

func (r *runner) taskRun(ctx context.Context, task types.Task) (needCancel bool) {
	var (
		currentTryTimes = 0
		err             error
	)
	r.logger.Infof("step %s started", task.GetName())
	if !r.waitingForRunning(ctx) {
		r.logger.Infof("job was finished, status=%s", r.GetStatus())
		return
	}

	task.SetStatue(types.RunningStatus)
	if err = r.updateTaskStatus(task); err != nil {
		r.logger.Errorf("update step status to running failed: %s", err)
		return
	}

	currentTryTimes += 1
	err = r.executor.DoOperation(ctx, task)
	if err != nil {
		r.logger.Warnf("do task %s failed: %s", task.GetName(), err)

		msg := fmt.Sprintf("task %s failed: %s", task.GetName(), err)
		task.SetStatue(types.FailedStatus)
		task.SetMessage(msg)
		if err = r.updateTaskStatus(task); err != nil {
			r.logger.Errorf("update task status to %s failed: %s", types.FailedStatus, err)
			return
		}

		needCancel = true
		err = r.pushEvent2FlowFSM(fsm.Event{Type: types.ExecuteFailedEvent, Status: r.GetStatus(), Message: msg, Obj: r.Flow})
		if err != nil {
			r.logger.Errorf("update job event failed %s", err)
		}
		return
	}

	r.logger.Infof("task %s succeed", task.GetName())
	task.SetStatue(types.SucceedStatus)
	if err = r.updateTaskStatus(task); err != nil {
		r.logger.Errorf("update task status to %s failed: %s", types.SucceedStatus, err)
		return
	}
	return
}

func (r *runner) initial() (err error) {
	if r.GetStatus() == "" {
		r.SetStatus(types.InitializingStatus)
		if err = r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
			r.logger.Errorf("initializing job status failed: %s")
			return err
		}
	}
	for _, task := range r.Tasks() {
		if task.GetStatue() == "" {
			task.SetStatue(types.InitializingStatus)
		}
	}
	if err = r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
		r.logger.Errorf("initializing step status failed: %s")
		return err
	}
	return
}

func (r *runner) nextBatch() ([]types.Task, error) {
	taskTowards := r.dag.nextBatchTasks()

	tasks := r.Tasks()
	nextBatch := make([]types.Task, 0, len(taskTowards))
	for _, t := range taskTowards {
		for i := range tasks {
			task := tasks[i]
			if task.GetName() == t.taskName {
				nextBatch = append(nextBatch, task)
				break
			}
		}
	}
	return nextBatch, nil
}

func (r *runner) waitingForRunning(ctx context.Context) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		default:
			switch r.GetStatus() {
			case types.SucceedStatus, types.FailedStatus, types.CanceledStatus, types.ErrorStatus:
				return false
			case types.InitializingStatus, types.RunningStatus:
				return true
			default:
				time.Sleep(time.Second * 15)
			}
		}
	}
}

func (r *runner) close(msg string) {
	r.logger.Debugf("stopping job: %s", msg)
	r.SetMessage(msg)

	if r.canF != nil {
		r.canF()
	}
}

func (r *runner) pushEvent2FlowFSM(event fsm.Event) error {
	err := r.fsm.Event(event)
	if err != nil {
		r.logger.Infof("push event to job FSM, event=%s, message=%s, error=%s", event.Type, event.Message, err)
		return err
	}
	return nil
}

func (r *runner) updateTaskStatus(task types.Task) error {
	if err := r.storage.SaveFlow(r.ctx, r.Flow); err != nil {
		return err
	}
	if IsFinishedStatus(task.GetStatue()) {
		r.dag.updateTaskStatus(task.GetName(), task.GetStatue())
	}
	return nil
}

func buildFlowFSM(r *runner) *fsm.FSM {
	m := fsm.New(fsm.Option{
		Obj:    r.Flow,
		Logger: r.logger.With("fsm"),
	})

	m.From([]string{types.InitializingStatus, types.RunningStatus}).
		To(types.RunningStatus).
		When(types.TriggerEvent).
		Do(r.handleJobRun)

	m.From([]string{types.RunningStatus}).
		To(types.SucceedStatus).
		When(types.ExecuteFinishEvent).
		Do(r.handleJobSucceed)

	m.From([]string{types.InitializingStatus, types.RunningStatus}).
		To(types.ErrorStatus).
		When(types.ExecuteErrorEvent).
		Do(r.handleJobFailed)

	m.From([]string{types.InitializingStatus, types.RunningStatus}).
		To(types.FailedStatus).
		When(types.ExecuteFailedEvent).
		Do(r.handleJobFailed)

	m.From([]string{types.InitializingStatus, types.PausedStatus}).
		To(types.CanceledStatus).
		When(types.ExecuteCancelEvent).
		Do(r.handleJobCancel)

	m.From([]string{types.RunningStatus}).
		To(types.PausedStatus).
		When(types.ExecutePauseEvent).
		Do(r.handleJobPause)

	m.From([]string{types.PausedStatus}).
		To(types.RunningStatus).
		When(types.ExecuteResumeEvent).
		Do(r.handleJobResume)

	return m
}
