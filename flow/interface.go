package flow

import (
	"github.com/zwwhdls/go-flow/fsm"
	"github.com/zwwhdls/go-flow/plugin"
)

type FID string

type Flow interface {
	fsm.Stateful
	ID() FID
	GetHooks() plugin.Hooks
	Setup(ctx *Context) error
	Teardown(ctx *Context)
	NextBatch(ctx *Context) ([]Task, error)
}

type TName string

type Task interface {
	fsm.Stateful
	Name() TName
	GetHooks() plugin.Hooks
	Setup(ctx *Context) error
	Do(ctx *Context)
	Teardown(ctx *Context)
}
