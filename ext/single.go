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

package ext

import (
	"fmt"
	"github.com/basenana/go-flow/flow"
	"github.com/basenana/go-flow/storage"
	"github.com/google/uuid"
)

type SingleFlow struct {
	*basic
	tasks []flow.Task
}

func (s SingleFlow) Type() flow.FType {
	return "Single"
}

var _ flow.Flow = &SingleFlow{}

func (s SingleFlow) GetHooks() flow.Hooks {
	return map[flow.HookType]flow.Hook{}
}

func (s SingleFlow) Setup(ctx *flow.Context) error {
	return nil
}

func (s SingleFlow) Teardown(ctx *flow.Context) {
	return
}

func (s *SingleFlow) NextBatch(ctx *flow.Context) ([]flow.Task, error) {
	if len(s.tasks) == 0 {
		return nil, nil
	}

	t := s.tasks[0]
	s.tasks = s.tasks[1:]

	return []flow.Task{t}, nil
}

func NewSingleFlow(s storage.Interface, tasks []flow.Task, policy flow.ControlPolicy) error {
	f := &SingleFlow{
		basic: &basic{
			id:     flow.FID(fmt.Sprintf("single-flow-%s", uuid.New().String())),
			status: flow.CreatingStatus,
			policy: policy,
		},
		tasks: tasks,
	}
	return s.SaveFlow(f)
}
