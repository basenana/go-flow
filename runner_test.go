package flow

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type logObserver struct {
	t *testing.T
}

func (l *logObserver) Handle(event UpdateEvent) {
	l.t.Logf("flow %s status: %s message: %s", event.Flow.ID, event.Flow.Status, event.Flow.Message)
	if event.Task != nil {
		l.t.Logf("flow %s task: %s status: %s message: %s", event.Flow.ID, event.Task.GetName(), event.Task.GetStatus(), event.Task.GetMessage())
	}
}

func TestSingleRunner(t *testing.T) {
	fb := NewFlowBuilder("test-f-01").Observer(&logObserver{t: t})

	fb.Task(NewFuncTask("t1", func(ctx context.Context) error {
		return nil
	}))
	fb.Task(NewFuncTask("t2", func(ctx context.Context) error {
		return nil
	}))
	fb.Task(NewFuncTask("t3", func(ctx context.Context) error {
		return nil
	}))

	f := fb.Finish()

	r := NewRunner(f)

	go func() {
		if err := r.Start(context.Background()); err != nil {
			t.Errorf("task failed: %s", err)
		}
	}()

	if err := eventualStatus(f, SucceedStatus, time.Second*10); err != nil {
		t.Errorf("wait flow status failed: %s", err)
		return
	}
}

func TestDAGRunner(t *testing.T) {
	fb := NewFlowBuilder("test-dag-01").
		Coordinator(NewDAGCoordinator()).
		Observer(&logObserver{t: t})

	fb.Task(WithDirector(NewFuncTask("dag-t1", func(ctx context.Context) error {
		return fmt.Errorf("mocked error")
	}), NextTask{
		OnSucceed: "dag-t2",
		OnFailed:  "dag-t3",
	}))
	fb.Task(NewFuncTask("dag-t2", func(ctx context.Context) error {
		return nil
	}))
	fb.Task(NewFuncTask("dag-t3", func(ctx context.Context) error {
		return nil
	}))
	fb.Task(NewFuncTask("dag-t4", func(ctx context.Context) error {
		return nil
	}))

	f := fb.Finish()

	r := NewRunner(f)

	go func() {
		if err := r.Start(context.Background()); err != nil {
			t.Errorf("task failed: %s", err)
		}
	}()

	if err := eventualStatus(f, SucceedStatus, time.Second*10); err != nil {
		t.Errorf("wait flow status failed: %s", err)
		return
	}
}

func TestPauseRunner(t *testing.T) {
	fb := NewFlowBuilder("test-pause-01").
		Coordinator(NewDAGCoordinator()).
		Observer(&logObserver{t: t})

	fb.Task(WithDirector(NewFuncTask("task-t1", func(ctx context.Context) error {
		return fmt.Errorf("mocked error")
	}), NextTask{
		OnSucceed: "task-t2",
		OnFailed:  "task-t3",
	}))
	fb.Task(NewFuncTask("task-t2", func(ctx context.Context) error {
		return nil
	}))
	fb.Task(NewFuncTask("task-t3", func(ctx context.Context) error {
		return nil
	}))
	fb.Task(NewFuncTask("task-t4", func(ctx context.Context) error {
		time.Sleep(time.Second * 5)
		return nil
	}))

	f := fb.Finish()
	r := NewRunner(f)

	go func() {
		if err := r.Start(context.Background()); err != nil {
			t.Errorf("task failed: %s", err)
		}
	}()

	if err := eventualStatus(f, RunningStatus, time.Second*10); err != nil {
		t.Errorf("wait flow status failed: %s", err)
		return
	}

	_ = r.Pause()
	if err := eventualStatus(f, PausedStatus, time.Second*10); err != nil {
		t.Errorf("wait flow status failed: %s", err)
		return
	}
	t.Logf("task paused %s", f.Status)

	_ = r.Resume()
	if err := eventualStatus(f, RunningStatus, time.Second*20); err != nil {
		t.Errorf("wait flow status failed: %s", err)
		return
	}
	t.Logf("task resumed %s", f.Status)

	if err := eventualStatus(f, SucceedStatus, time.Second*10); err != nil {
		t.Errorf("wait flow status failed: %s", err)
		return
	}
}

func eventualStatus(f *Flow, status string, timeout time.Duration) error {
	start := time.Now()
	for {
		if time.Since(start) > timeout {
			return fmt.Errorf("expect %s, but got %s", status, f.Status)
		}

		if f.Status == status {
			return nil
		}
		time.Sleep(time.Millisecond)
	}
}
