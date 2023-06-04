/*
   Copyright 2022 Go-Flow Authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package storage

import (
	"context"
	"fmt"
	"github.com/basenana/go-flow/flow"
	"sync"
)

const (
	flowKeyTpl = "storage.flow.%s"
	taskKeyTpl = "storage.flow.%s.task.%s"
)

var ErrNotFound = fmt.Errorf("not found")

type MemStorage struct {
	sync.Map
}

func (m *MemStorage) GetFlow(ctx context.Context, flowId string) (*flow.Flow, error) {
	k := fmt.Sprintf(flowKeyTpl, flowId)
	obj, ok := m.Load(k)
	if !ok {
		return nil, ErrNotFound
	}
	return obj.(*flow.Flow), nil
}

func (m *MemStorage) SaveFlow(ctx context.Context, flow *flow.Flow) error {
	k := fmt.Sprintf(flowKeyTpl, flow.ID)
	m.Store(k, flow)
	return nil
}

func (m *MemStorage) DeleteFlow(ctx context.Context, flowId string) error {
	k := fmt.Sprintf(flowKeyTpl, flowId)
	_, ok := m.LoadAndDelete(k)
	if !ok {
		return ErrNotFound
	}
	return nil
}

func (m *MemStorage) SaveTask(ctx context.Context, flowId string, task *flow.Task) error {
	k := fmt.Sprintf(taskKeyTpl, flowId, task.Name)
	m.Store(k, task)
	return nil
}

func NewInMemoryStorage() Interface {
	return &MemStorage{}
}
