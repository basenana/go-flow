package flow

import (
	"go-flow/context"
	"go-flow/fsm"
	"go-flow/task"
)

type FID string

type Flow interface {
	fsm.Stateful
	ID() FID
	Setup(ctx context.FlowContext)
	Teardown(ctx context.FlowContext)
	NextBatch(ctx context.FlowContext) ([]task.Task, error)
}
