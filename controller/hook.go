package controller

import (
	"github.com/zwwhdls/go-flow/flow"
)

func flowHookDecorator(ctx *flow.Context, f flow.Flow, h flow.Hook) func(args ...interface{}) error {
	return func(args ...interface{}) error {
		return h(ctx, f, nil)
	}
}

func taskHookDecorator(ctx *flow.Context, f flow.Flow, t flow.Task, h flow.Hook) func(args ...interface{}) error {
	return func(args ...interface{}) error {
		return h(ctx, f, t)
	}
}
