package storage

import (
	"github.com/zwwhdls/go-flow/flow"
)

type Interface interface {
	GetFlow(flowId flow.FID) (flow.Flow, error)
	GetFlowMeta(flowId flow.FID) (*FlowMeta, error)
	SaveFlow(flow flow.Flow) error
	DeleteFlow(flowId flow.FID) (flow.Flow, error)
	SaveTask(flowId flow.FID, task flow.Task) error
	DeleteTask(flowId flow.FID, taskName flow.TName) (flow.Task, error)
}
