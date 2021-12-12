package controller

import (
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/plugin"
)

func flowHookDecorator(ctx *flow.Context, f flow.Flow, h plugin.Hook) func(args ...interface{}) error {
	return func(args ...interface{}) error {
		return h(ctx, f, nil)
	}
}

func taskHookDecorator(ctx *flow.Context, f flow.Flow, t flow.Task, h plugin.Hook) func(args ...interface{}) error {
	return func(args ...interface{}) error {
		return h(ctx, f, t)
	}
}
