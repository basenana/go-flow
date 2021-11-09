package storage

import (
	"go-flow/flow"
)

type Interface interface {
	GetFlow(flowId flow.FID) (flow.Flow, error)
	SaveFlow(flow flow.Flow) error
	GetTask(flowId flow.FID, taskName flow.TName) (flow.Task, error)
	SaveTask(task flow.Task) error
}
