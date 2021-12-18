package flow

import (
	"github.com/zwwhdls/go-flow/fsm"
)

type FID string

type Flow interface {
	fsm.Stateful
	ID() FID
	GetHooks() Hooks
	Setup(ctx *Context) error
	Teardown(ctx *Context)
	NextBatch(ctx *Context) ([]Task, error)
	GetControlPolicy() ControlPolicy
}

type TName string

type Task interface {
	fsm.Stateful
	Name() TName
	Setup(ctx *Context) error
	Do(ctx *Context) error
	Teardown(ctx *Context)
}
