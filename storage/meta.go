package storage

import (
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
)

type FlowMeta struct {
	Type       flow.FType
	Id         flow.FID
	Status     fsm.Status
	TaskStatus map[flow.TName]fsm.Status
}

func (f FlowMeta) QueryTaskStatus(taskName flow.TName) (fsm.Status, error) {
	status, ok := f.TaskStatus[taskName]
	if !ok {
		return "", NotFound
	}
	return status, nil
}
