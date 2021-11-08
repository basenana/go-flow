package task

import (
	"go-flow/context"
	"go-flow/fsm"
)

type TName string

type Task interface {
	fsm.Stateful
	Name() TName
	Setup(ctx context.TaskContext)
	Teardown(ctx context.TaskContext)
}
