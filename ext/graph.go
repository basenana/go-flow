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
	"errors"
	"fmt"
	"github.com/basenana/go-flow/flow"
	"github.com/basenana/go-flow/storage"
	"github.com/google/uuid"
)

type GraphFlow struct {
	*basic
	tasks   map[string]*GraphTask
	batches [][]string
}

func (g *GraphFlow) Setup(ctx *flow.Context) error {
	ctx.Succeed()
	return nil
}

func (g *GraphFlow) Teardown(ctx *flow.Context) {
	return
}

func (g *GraphFlow) taskDoFunc(ctx *flow.Context, appId string) error {
	// define your task func here
	ctx.Logger.Infof("app %s task done", appId)
	ctx.Succeed()
	return nil
}

func (g *GraphFlow) taskTearDownFunc(ctx *flow.Context, appId string) {
	// define your task func here
	ctx.Logger.Infof("app %s task teardown", appId)
	ctx.Succeed()
}

func (g *GraphFlow) taskSetUpFunc(ctx *flow.Context, appId string) error {
	// define your task func here
	ctx.Logger.Infof("app %s task setup", appId)
	ctx.Succeed()
	return nil
}

func (g *GraphFlow) NextBatch(ctx *flow.Context) ([]flow.ITask, error) {
	if len(g.batches) == 0 {
		return nil, nil
	}

	crtBatch := g.batches[0]
	g.batches = g.batches[1:]

	tasks := make([]flow.ITask, len(crtBatch))
	for i, tName := range crtBatch {
		task := g.tasks[tName]
		task.doFunc = func(ctx *flow.Context) error {
			return g.taskDoFunc(ctx, task.AppId)
		}
		task.setupFunc = func(ctx *flow.Context) error {
			return g.taskSetUpFunc(ctx, task.AppId)
		}
		task.teardownFunc = func(ctx *flow.Context) {
			g.taskTearDownFunc(ctx, task.AppId)
		}
		tasks[i] = task
	}
	return tasks, nil
}

func NewGraphFlow(s storage.Interface, tasks []*GraphTask, dep *TaskDep, policy flow.ControlPolicy) (*GraphFlow, error) {
	if dep == nil || dep.taskEdges == nil {
		dep = NewTaskDep()
	}

	batches, err := dep.Lists()
	if err != nil {
		return nil, err
	}

	taskMap := make(map[string]*GraphTask)
	for i, t := range tasks {
		taskMap[t.TaskName] = tasks[i]
	}

	f := &GraphFlow{
		basic: &basic{
			id:     fmt.Sprintf("graph-flow-%s", uuid.New().String()),
			policy: policy,
		},
		tasks:   taskMap,
		batches: batches,
	}
	return f, s.SaveFlow(f)
}

type GraphTask struct {
	TaskName string
	AppId    string
	Message  string
	status   string

	setupFunc    func(ctx *flow.Context) error
	doFunc       func(ctx *flow.Context) error
	teardownFunc func(ctx *flow.Context)
}

func (g GraphTask) GetStatus() string {
	return g.status
}

func (g *GraphTask) SetStatus(status string) {
	g.status = status
}

func (g GraphTask) GetMessage() string {
	return g.Message
}

func (g *GraphTask) SetMessage(msg string) {
	g.Message = msg
}

func (g GraphTask) Name() string {
	return g.TaskName
}

func (g GraphTask) Setup(ctx *flow.Context) error {
	return g.setupFunc(ctx)
}

func (g GraphTask) Do(ctx *flow.Context) error {
	return g.doFunc(ctx)
}

func (g GraphTask) Teardown(ctx *flow.Context) {
	g.teardownFunc(ctx)
}

func NewGraphTask(name, appId, msg string) *GraphTask {
	return &GraphTask{
		TaskName: name,
		AppId:    appId,
		Message:  msg,
		status:   flow.CreatingStatus,
	}
}

type TaskDep struct {
	taskSet   TaskNameSet
	taskEdges map[string][]string
	preCount  map[string]int
}

func (t *TaskDep) RunOrder(firstTask string, tasks []string) {
	t.taskSet.Insert(firstTask)
	for _, task := range tasks {
		t.taskSet.Insert(task)

		edges, ok := t.taskEdges[firstTask]
		if !ok {
			edges = make([]string, 0)
		}

		exist := false
		for _, d := range edges {
			if d == task {
				exist = true
				break
			}
		}
		if exist {
			continue
		}
		edges = append(edges, task)
		t.taskEdges[firstTask] = edges
		t.preCount[task] += 1
	}
}

func (t *TaskDep) Lists() (result [][]string, err error) {

	for t.taskSet.Len() != 0 {
		batch := make([]string, 0)
		duty := NewTaskNameSet()
		for _, tName := range t.taskSet.List() {
			if duty.Has(tName) {
				continue
			}

			if t.preCount[tName] > 0 {
				continue
			}
			for _, nextTask := range t.taskEdges[tName] {
				t.preCount[nextTask] -= 1
				duty.Insert(nextTask)
			}
			batch = append(batch, tName)
			t.taskSet.Del(tName)
		}

		if t.taskSet.Len() > 0 && len(batch) == 0 {
			return nil, errors.New("task graph has a loop")
		}

		result = append(result, batch)
	}

	return
}

func NewTaskDep() *TaskDep {
	return &TaskDep{
		taskSet:   NewTaskNameSet(),
		taskEdges: map[string][]string{},
		preCount:  map[string]int{},
	}
}
