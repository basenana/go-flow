/*
   Copyright 2024 Go-Flow Authors

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
	"sync"
	"time"
)

func NewRunner(f *Flow) *Runner {
	return &Runner{Flow: f}
}

type Runner struct {
	*Flow

	ctx     context.Context
	canF    context.CancelFunc
	fsm     *FSM
	stopCh  chan struct{}
	started bool
	mux     sync.Mutex
}

func (r *Runner) Start(ctx context.Context) (err error) {
	if IsFinishedStatus(r.Status) {
		return
	}

	r.ctx, r.canF = context.WithCancel(ctx)
	if err = r.executor.Setup(ctx); err != nil {
		r.SetStatus(ErrorStatus, err.Error())
		return err
	}

	r.fsm = buildFlowFSM(r)
	r.stopCh = make(chan struct{})
	if !r.waitingForRunning(ctx) {
		return
	}

	if err = r.pushEvent2FlowFSM(statusEvent{Type: TriggerEvent}); err != nil {
		r.SetStatus(ErrorStatus, err.Error())
		return err
	}

	// waiting all task down
	<-r.stopCh

	return r.executor.Teardown(ctx)
}

func (r *Runner) Pause() error {
	if r.Status == RunningStatus {
		return r.pushEvent2FlowFSM(statusEvent{Type: ExecutePauseEvent})
	}
	return nil
}

func (r *Runner) Resume() error {
	if r.Status == PausedStatus {
		return r.pushEvent2FlowFSM(statusEvent{Type: ExecuteResumeEvent})
	}
	return nil
}

func (r *Runner) Cancel() error {
	if IsFinishedStatus(r.Status) {
		return nil
	}
	return r.pushEvent2FlowFSM(statusEvent{Type: ExecuteCancelEvent})
}

func (r *Runner) handleJobRun(event statusEvent) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.started {
		return nil
	}
	r.started = true
	r.SetStatus(RunningStatus, event.Message)

	go func() {
		var (
			isFinish bool
			err      error
		)
		defer func() {
			close(r.stopCh)
		}()

		for {
			select {
			case <-r.ctx.Done():
				err = r.ctx.Err()
			default:
				isFinish, err = r.batchRun()
			}

			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				_ = r.pushEvent2FlowFSM(statusEvent{Type: ExecuteErrorEvent, Message: err.Error()})
				return
			}
			if isFinish {
				if !IsFinishedStatus(r.Status) {
					_ = r.pushEvent2FlowFSM(statusEvent{Type: ExecuteFinishEvent, Message: "finish"})
				}
				return
			}
		}
	}()
	return nil
}

func (r *Runner) handleJobPause(event statusEvent) error {
	r.SetStatus(PausedStatus, event.Message)
	return nil
}

func (r *Runner) handleJobResume(event statusEvent) error {
	r.SetStatus(RunningStatus, event.Message)
	return nil
}

func (r *Runner) handleJobSucceed(event statusEvent) error {
	r.SetStatus(SucceedStatus, event.Message)
	r.close()
	return nil
}

func (r *Runner) handleJobFailed(event statusEvent) error {
	r.SetStatus(FailedStatus, event.Message)
	r.close()
	return nil
}

func (r *Runner) handleJobCancel(event statusEvent) error {
	r.SetStatus(CanceledStatus, event.Message)
	r.close()
	return nil
}

func (r *Runner) batchRun() (finish bool, err error) {
	if !r.waitingForRunning(r.ctx) {
		return true, nil
	}

	defer func() {
		if panicErr := doRecover(); panicErr != nil {
			err = panicErr
		}
	}()

	var batch []Task
	batch, err = r.coordinator.NextBatch(r.ctx)
	if err != nil {
		return
	}

	if len(batch) == 0 {
		return true, nil
	}

	batchCtx, batchCanF := context.WithCancel(r.ctx)
	defer batchCanF()
	wg := sync.WaitGroup{}
	for i := range batch {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			if needCancel := r.taskRun(batchCtx, t); needCancel {
				batchCanF()
			}
		}(batch[i])
	}
	wg.Wait()

	return false, nil
}

func (r *Runner) taskRun(ctx context.Context, task Task) (needCancel bool) {
	var (
		currentTryTimes = 0
		err             error
	)
	if !r.waitingForRunning(ctx) {
		return
	}

	r.SetTaskStatue(task, RunningStatus, "")
	currentTryTimes += 1
	err = r.executor.Exec(ctx, r.Flow, task)
	if err != nil {
		msg := fmt.Sprintf("task %s failed: %s", task.GetName(), err)

		op := r.coordinator.HandleFail(task, err)
		switch op {
		case FailAndInterrupt:
			r.SetTaskStatue(task, FailedStatus, msg)
			_ = r.pushEvent2FlowFSM(statusEvent{Type: ExecuteFailedEvent, Message: msg})
		case FailAndPause:
			r.SetTaskStatue(task, PausedStatus, msg)
			_ = r.pushEvent2FlowFSM(statusEvent{Type: ExecutePauseEvent, Message: msg})
		case FailButContinue:
			r.SetTaskStatue(task, FailedStatus, msg)
		}

		needCancel = true
		return
	}

	r.SetTaskStatue(task, SucceedStatus, "")
	return
}

func (r *Runner) waitingForRunning(ctx context.Context) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		default:
			switch r.Status {
			case SucceedStatus, FailedStatus, CanceledStatus, ErrorStatus:
				return false
			case InitializingStatus, RunningStatus:
				return true
			default:
				time.Sleep(time.Second * 15)
			}
		}
	}
}

func (r *Runner) close() {
	if r.canF != nil {
		r.canF()
	}
}

func (r *Runner) pushEvent2FlowFSM(event statusEvent) error {
	err := r.fsm.Event(event)
	if err != nil {
		return err
	}
	return nil
}

func buildFlowFSM(r *Runner) *FSM {
	m := NewFSM(r.Status)

	m.From([]string{InitializingStatus, RunningStatus}).
		To(RunningStatus).
		When(TriggerEvent).
		Do(r.handleJobRun)

	m.From([]string{RunningStatus}).
		To(SucceedStatus).
		When(ExecuteFinishEvent).
		Do(r.handleJobSucceed)

	m.From([]string{InitializingStatus, RunningStatus}).
		To(ErrorStatus).
		When(ExecuteErrorEvent).
		Do(r.handleJobFailed)

	m.From([]string{InitializingStatus, RunningStatus}).
		To(FailedStatus).
		When(ExecuteFailedEvent).
		Do(r.handleJobFailed)

	m.From([]string{InitializingStatus, RunningStatus, PausedStatus}).
		To(CanceledStatus).
		When(ExecuteCancelEvent).
		Do(r.handleJobCancel)

	m.From([]string{RunningStatus}).
		To(PausedStatus).
		When(ExecutePauseEvent).
		Do(r.handleJobPause)

	m.From([]string{PausedStatus}).
		To(RunningStatus).
		When(ExecuteResumeEvent).
		Do(r.handleJobResume)

	return m
}
