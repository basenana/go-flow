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

package go_flow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Runner interface {
	Start(ctx context.Context) error
	Pause() error
	Resume() error
	Cancel() error
}

func NewRunner(f *Flow) Runner {
	return &runner{Flow: f}
}

type runner struct {
	*Flow

	ctx     context.Context
	canF    context.CancelFunc
	fsm     *FSM
	stopCh  chan struct{}
	started bool
	mux     sync.Mutex
}

func (r *runner) Start(ctx context.Context) (err error) {
	if IsFinishedStatus(r.Status) {
		return
	}

	r.ctx, r.canF = context.WithCancel(ctx)
	if err = r.executor.Setup(ctx); err != nil {
		r.SetStatus(ErrorStatus)
		r.SetMessage(err.Error())
		return err
	}

	r.fsm = buildFlowFSM(r)
	r.stopCh = make(chan struct{})
	if !r.waitingForRunning(ctx) {
		return
	}

	if err = r.pushEvent2FlowFSM(Event{Type: TriggerEvent, Status: r.Flow.GetStatus()}); err != nil {
		r.SetStatus(ErrorStatus)
		r.SetMessage(err.Error())
		return err
	}

	// waiting all task down
	<-r.stopCh

	return r.executor.Teardown(ctx)
}

func (r *runner) Pause() error {
	if r.Status == RunningStatus {
		return r.pushEvent2FlowFSM(Event{Type: ExecutePauseEvent})
	}
	return nil
}

func (r *runner) Resume() error {
	if r.Status == PausedStatus {
		return r.pushEvent2FlowFSM(Event{Type: ExecuteResumeEvent})
	}
	return nil
}

func (r *runner) Cancel() error {
	if IsFinishedStatus(r.Status) {
		return nil
	}
	return r.pushEvent2FlowFSM(Event{Type: ExecuteCancelEvent})
}

func (r *runner) handleJobRun(event Event) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	if r.started {
		return nil
	}
	r.started = true
	r.SetStatus(RunningStatus)

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
				_ = r.pushEvent2FlowFSM(Event{Type: ExecuteErrorEvent, Status: r.GetStatus(), Message: err.Error()})
				return
			}
			if isFinish {
				if !IsFinishedStatus(r.GetStatus()) {
					_ = r.pushEvent2FlowFSM(Event{Type: ExecuteFinishEvent, Status: r.GetStatus(), Message: "finish"})
				}
				return
			}
		}
	}()
	return nil
}

func (r *runner) handleJobPause(event Event) error {
	return nil
}

func (r *runner) handleJobResume(event Event) error {
	return nil
}

func (r *runner) handleJobSucceed(event Event) error {
	r.close("succeed")
	return nil
}

func (r *runner) handleJobFailed(event Event) error {
	r.close(event.Message)
	return nil
}

func (r *runner) handleJobCancel(event Event) error {
	r.close(event.Message)
	return nil
}

func (r *runner) batchRun() (finish bool, err error) {
	if !r.waitingForRunning(r.ctx) {
		return true, nil
	}

	defer func() {
		if panicErr := doRecover(); panicErr != nil {
			err = panicErr
		}
	}()

	var batch []Task
	if batch, err = r.nextBatch(); err != nil {
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

func (r *runner) taskRun(ctx context.Context, task Task) (needCancel bool) {
	var (
		currentTryTimes = 0
		err             error
	)
	if !r.waitingForRunning(ctx) {
		return
	}

	r.SetTaskStatue(task, RunningStatus, "")
	currentTryTimes += 1
	_, err = r.executor.Exec(ctx, r.Flow, task)
	if err != nil {
		msg := fmt.Sprintf("task %s failed: %s", task.GetName(), err)
		r.SetTaskStatue(task, FailedStatus, msg)

		needCancel = true
		_ = r.pushEvent2FlowFSM(Event{Type: ExecuteFailedEvent, Status: r.GetStatus(), Message: msg})
		return
	}

	r.SetTaskStatue(task, SucceedStatus, "")
	return
}

func (r *runner) nextBatch() ([]Task, error) {
	nextBatch, err := r.coordinator.NextBatch(r.ctx)
	if err != nil {
		return nil, err
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

func (r *runner) close(msg string) {
	r.SetMessage(msg)

	if r.canF != nil {
		r.canF()
	}
}

func (r *runner) pushEvent2FlowFSM(event Event) error {
	err := r.fsm.Event(event)
	if err != nil {
		return err
	}
	return nil
}

func buildFlowFSM(r *runner) *FSM {
	m := NewFSM()

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

	m.From([]string{InitializingStatus, PausedStatus}).
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
