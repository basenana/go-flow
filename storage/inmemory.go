package storage

import (
	"fmt"
	"github.com/zwwhdls/go-flow/flow"
	"sync"
)

const (
	flowKeyTpl     = "storage.flow.%s"
	flowMetaKeyTpl = "storage.flow-meta.%s"
	taskKeyTpl     = "storage.flow.%s.task.%s"
)

var ErrNotFound = fmt.Errorf("not found")

type MemStorage struct {
	sync.Map
}

func (m *MemStorage) GetFlow(flowId flow.FID) (flow.Flow, error) {
	k := fmt.Sprintf(flowKeyTpl, flowId)
	obj, ok := m.Load(k)
	if !ok {
		return nil, ErrNotFound
	}
	return obj.(flow.Flow), nil
}

func (m *MemStorage) GetFlowMeta(flowId flow.FID) (*FlowMeta, error) {
	k := fmt.Sprintf(flowMetaKeyTpl, flowId)
	obj, ok := m.Load(k)
	if !ok {
		return nil, ErrNotFound
	}
	return obj.(*FlowMeta), nil
}

func (m *MemStorage) SaveFlow(flow flow.Flow) error {
	mk := fmt.Sprintf(flowMetaKeyTpl, flow.ID())
	_, ok := m.Load(mk)
	if !ok {
		m.Store(mk, &FlowMeta{})
	}

	k := fmt.Sprintf(flowKeyTpl, flow.ID())
	m.Store(k, flow)
	return nil
}

func (m *MemStorage) DeleteFlow(flowId flow.FID) error {
	k := fmt.Sprintf(flowKeyTpl, flowId)
	_, ok := m.LoadAndDelete(k)
	if !ok {
		return ErrNotFound
	}
	return nil
}

func (m *MemStorage) SaveTask(flowId flow.FID, task flow.Task) error {
	k := fmt.Sprintf(taskKeyTpl, flowId, task.Name())
	m.Store(k, task)
	return nil
}

func (m *MemStorage) DeleteTask(flowId flow.FID, taskName flow.TName) error {
	k := fmt.Sprintf(taskKeyTpl, flowId, taskName)
	_, ok := m.LoadAndDelete(k)
	if !ok {
		return ErrNotFound
	}
	return nil
}

func NewInMemoryStorage() Interface {
	return &MemStorage{}
}
