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
	coor := NewDAGCoordinator()
	builder := NewFlowBuilder("dag-flow-01").Coordinator(coor)

	task8 := NewFuncTask("task-8", NT("task-8"))
	task7 := NewFuncTask("task-7", NT("task-7"))
	task6 := NewFuncTask("task-6", NT("task-6"))
	task5 := NewFuncTask("task-5", NT("task-5"))
	task4 := NewFuncTask("task-4", NT("task-4"))
	task3 := NewFuncTask("task-3", NT("task-3"))
	task2 := NewFuncTask("task-2", NT("task-2"))
	task1 := NewFuncTask("task-1", NT("task-1"))

	builder.Task(WithDependent(task8, []string{"task-5", "task-4"}))
	builder.Task(WithDependent(task7, []string{"task-6"}))
	builder.Task(WithDependent(task6, []string{"task-5", "task-4", "task-3"}))
	builder.Task(task5)
	builder.Task(WithDependent(task3, []string{"task-1"}))
	builder.Task(WithDependent(task4, []string{"task-2"}))
	builder.Task(task2)
	builder.Task(task1)

	dagFlow := builder.Finish()

	r := NewRunner(dagFlow)

	go func() {
		if err := r.Start(context.Background()); err != nil {
			t.Errorf("task failed: %s", err)
		}
	}()

	if err := eventualStatus(dagFlow, SucceedStatus, time.Second*10); err != nil {
		t.Errorf("wait flow status failed: %s", err)
		return
	}
}

func TestPauseRunner(t *testing.T) {
	fb := NewFlowBuilder("test-pause-01").
		Coordinator(NewDAGCoordinator()).
		Observer(&logObserver{t: t})

	fb.Task(WithDependent(NewFuncTask("task-t1", func(ctx context.Context) error {
		return nil
	}), []string{"task-t2", "task-t3"}))
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
			return fmt.Errorf("expect %s, but got %s: %s", status, f.Status, f.Message)
		}

		if f.Status == status {
			return nil
		}
		time.Sleep(time.Millisecond)
	}
}

func NT(name string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		fmt.Printf("task %s running\n", name)
		time.Sleep(time.Second)
		return nil
	}
}
