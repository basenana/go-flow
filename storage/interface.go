package storage

import (
	"go-flow/flow"
	"go-flow/task"
)

type Interface interface {
	GetFlow(flowId flow.FID) (flow.Flow, error)
	SaveFlow(flow flow.Flow) error
	GetTask(flowId flow.FID, taskName task.TName) (task.Task, error)
	SaveTask(task task.Task) error
}
