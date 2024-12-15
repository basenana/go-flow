/*
   Copyright 2024 Go-Flow Authors

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

package go_flow

type Flow struct {
	ID      string
	Status  string
	Message string

	tasks       []Task
	executor    Executor
	coordinator Coordinator
	observer    []Observer
}

func (f *Flow) GetStatus() string {
	return f.Status
}

func (f *Flow) SetStatus(status string) {
	f.Status = status
}

func (f *Flow) GetMessage() string {
	return f.Message
}

func (f *Flow) SetMessage(msg string) {
	f.Message = msg
}

func (f *Flow) SetTaskStatue(task Task, status, msg string) {
}

func (f *Flow) dispatch(event Event) {
	for _, ob := range f.observer {
		ob.Handle(event)
	}
	return
}

func NewFlowBuilder(id string) *Builder {
	return &Builder{id: id, executor: &simpleExecutor{}, coordinator: &pipelineCoordinator{}}
}

type Builder struct {
	id          string
	tasks       []Task
	executor    Executor
	coordinator Coordinator
	observer    []Observer
}

func (b *Builder) Task(task Task) *Builder {
	b.tasks = append(b.tasks, task)
	return b
}

func (b *Builder) Executor(executor Executor) *Builder {
	b.executor = executor
	return b
}

func (b *Builder) Coordinator(coordinator Coordinator) *Builder {
	b.coordinator = coordinator
	return b
}

func (b *Builder) Observer(observer Observer) *Builder {
	b.observer = append(b.observer, observer)
	return b
}

func (b *Builder) Finish() *Flow {
	f := &Flow{
		ID:          b.id,
		Status:      InitializingStatus,
		tasks:       make([]Task, 0, len(b.tasks)),
		executor:    b.executor,
		coordinator: b.coordinator,
		observer:    make([]Observer, 0, len(b.observer)),
	}

	for i := range b.tasks {
		f.tasks = append(f.tasks, b.tasks[i])
		f.coordinator.NewTask(b.tasks[i])
	}
	for i := range b.observer {
		f.observer = append(f.observer, b.observer[i])
	}
	return f
}
