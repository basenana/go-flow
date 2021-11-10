package flow

import (
	"github.com/zwwhdls/go-flow/fsm"
)

type FID string

type Flow interface {
	fsm.Stateful
	ID() FID
	Setup(ctx FlowContext)
	Teardown(ctx FlowContext)
	NextBatch(ctx FlowContext) ([]Task, error)
}

type TName string

type Task interface {
	fsm.Stateful
	Name() TName
	Setup(ctx TaskContext)
	Do(ctx TaskContext)
	Teardown(ctx TaskContext)
}
