package flow

import (
	"github.com/zwwhdls/go-flow/context"
	"github.com/zwwhdls/go-flow/fsm"
)

type FID string

type Flow interface {
	fsm.Stateful
	ID() FID
	Setup(ctx context.FlowContext)
	Teardown(ctx context.FlowContext)
	NextBatch(ctx context.FlowContext) ([]Task, error)
}

type TName string

type Task interface {
	fsm.Stateful
	Name() TName
	Setup(ctx context.TaskContext)
	Do(ctx context.TaskContext)
	Teardown(ctx context.TaskContext)
}
