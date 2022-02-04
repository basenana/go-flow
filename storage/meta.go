package storage

import (
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
)

type FlowMeta struct{}

func (f FlowMeta) QueryTaskStatus(taskName flow.TName) (fsm.Status, error) {
	return "", ErrNotFound
}
