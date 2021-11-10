package controller

import (
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
	"github.com/zwwhdls/go-flow/plugin"
)

type hookWarp struct {
	WhenTrigger        fsm.Handler
	WhenInitFinish     fsm.Handler
	WhenExecuteSucceed fsm.Handler
	WhenExecuteFailed  fsm.Handler
	WhenExecutePause   fsm.Handler
	WhenExecuteResume  fsm.Handler
	WhenExecuteCancel  fsm.Handler

	WhenTaskTrigger        fsm.Handler
	WhenTaskInitFinish     fsm.Handler
	WhenTaskInitFailed     fsm.Handler
	WhenTaskExecuteSucceed fsm.Handler
	WhenTaskExecuteFailed  fsm.Handler
	WhenTaskExecutePause   fsm.Handler
	WhenTaskExecuteResume  fsm.Handler
	WhenTaskExecuteCancel  fsm.Handler
}

func buildHookWithDefault(ctx flow.FlowContext, f flow.Flow, hook plugin.Hook) hookWarp {
	return hookWarp{
		WhenTrigger: func(args ...interface{}) error {
			if hook.WhenTrigger != nil {
				return hook.WhenTrigger(ctx, f)
			}
			return nil
		},
		WhenInitFinish: func(args ...interface{}) error {
			if hook.WhenInitFinish != nil {
				return hook.WhenInitFinish(ctx, f)
			}
			return nil
		},
		WhenExecuteSucceed: func(args ...interface{}) error {
			if hook.WhenExecuteSucceed != nil {
				return hook.WhenExecuteSucceed(ctx, f)
			}
			return nil
		},
		WhenExecuteFailed: func(args ...interface{}) error {
			if hook.WhenExecuteFailed != nil {
				return hook.WhenExecuteFailed(ctx, f)
			}
			return nil
		},
		WhenExecutePause: func(args ...interface{}) error {
			if hook.WhenExecutePause != nil {
				return hook.WhenExecutePause(ctx, f)
			}
			return nil
		},
		WhenExecuteResume: func(args ...interface{}) error {
			if hook.WhenExecuteResume != nil {
				return hook.WhenExecuteResume(ctx, f)
			}
			return nil
		},
		WhenExecuteCancel: func(args ...interface{}) error {
			if hook.WhenExecuteCancel != nil {
				return hook.WhenExecuteCancel(ctx, f)
			}
			return nil
		},
	}
}

func buildTaskHookWithDefault(ctx flow.TaskContext, t flow.Task, hook plugin.Hook) hookWarp {
	return hookWarp{
		WhenTaskTrigger: func(args ...interface{}) error {
			if hook.WhenTaskTrigger != nil {
				return hook.WhenTaskTrigger(ctx, t)
			}
			return nil
		},
		WhenTaskInitFinish: func(args ...interface{}) error {
			if hook.WhenTaskInitFinish != nil {
				return hook.WhenTaskInitFinish(ctx, t)
			}
			return nil
		},
		WhenTaskExecuteSucceed: func(args ...interface{}) error {
			if hook.WhenTaskExecuteSucceed != nil {
				return hook.WhenTaskExecuteSucceed(ctx, t)
			}
			return nil
		},
		WhenTaskExecuteFailed: func(args ...interface{}) error {
			if hook.WhenTaskExecuteFailed != nil {
				return hook.WhenTaskExecuteFailed(ctx, t)
			}
			return nil
		},
		WhenTaskExecutePause: func(args ...interface{}) error {
			if hook.WhenTaskExecutePause != nil {
				return hook.WhenTaskExecutePause(ctx, t)
			}
			return nil
		},
		WhenTaskExecuteResume: func(args ...interface{}) error {
			if hook.WhenTaskExecuteResume != nil {
				return hook.WhenTaskExecuteResume(ctx, t)
			}
			return nil
		},
		WhenTaskExecuteCancel: func(args ...interface{}) error {
			if hook.WhenTaskExecuteCancel != nil {
				return hook.WhenTaskExecuteCancel(ctx, t)
			}
			return nil
		},
	}
}
