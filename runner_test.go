package flow

import (
	"context"
	"fmt"
	"testing"
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

	if err := r.Start(context.Background()); err != nil {
		t.Errorf("task failed: %s", err)
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

	if err := r.Start(context.Background()); err != nil {
		t.Errorf("task failed: %s", err)
	}
}
