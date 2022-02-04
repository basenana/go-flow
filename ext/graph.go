package ext

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/storage"
)

type GraphFlow struct {
	*basic
	tasks   map[flow.TName]flow.Task
	batches [][]flow.TName
}

var _ flow.Flow = &GraphFlow{}

func (g GraphFlow) Type() flow.FType {
	return "GraphFlow"
}

func (g GraphFlow) GetHooks() flow.Hooks {
	return map[flow.HookType]flow.Hook{}
}

func (g GraphFlow) Setup(ctx *flow.Context) error {
	return nil
}

func (g GraphFlow) Teardown(ctx *flow.Context) {
	return
}

func (g GraphFlow) NextBatch(ctx *flow.Context) ([]flow.Task, error) {
	if len(g.batches) == 0 {
		return nil, nil
	}

	crtBatch := g.batches[0]
	g.batches = g.batches[1:]

	tasks := make([]flow.Task, len(crtBatch))
	for i, tName := range crtBatch {
		tasks[i] = g.tasks[tName]
	}
	return tasks, nil
}

func NewGraphFlow(s storage.Interface, tasks []flow.Task, dep *TaskDep, policy flow.ControlPolicy) error {
	if dep == nil || dep.taskEdges == nil {
		dep = NewTaskDep()
	}

	batches, err := dep.Lists()
	if err != nil {
		return err
	}

	taskMap := make(map[flow.TName]flow.Task)
	for i, t := range tasks {
		taskMap[t.Name()] = tasks[i]
	}

	f := &GraphFlow{
		basic: &basic{
			id:     flow.FID(fmt.Sprintf("graph-flow-%s", uuid.New().String())),
			status: flow.CreatingStatus,
			policy: policy,
		},
		tasks:   taskMap,
		batches: batches,
	}
	return s.SaveFlow(f)
}

type TaskDep struct {
	taskSet   TaskNameSet
	taskEdges map[flow.TName][]flow.TName
	preCount  map[flow.TName]int
}

func (t *TaskDep) RunOrder(firstTask, secondTask flow.TName) {
	t.taskSet.Insert(firstTask)
	t.taskSet.Insert(secondTask)

	edges, ok := t.taskEdges[firstTask]
	if !ok {
		edges = make([]flow.TName, 0)
	}

	for _, d := range edges {
		if d == secondTask {
			return
		}
	}

	edges = append(edges, secondTask)
	t.taskEdges[firstTask] = edges
	t.preCount[firstTask] += 1
}

func (t *TaskDep) Lists() (result [][]flow.TName, err error) {

	for t.taskSet.Len() == 0 {
		batch := make([]flow.TName, 0)

		for _, tName := range t.taskSet.List() {
			if t.preCount[tName] > 0 {
				continue
			}
			for _, nextTask := range t.taskEdges[tName] {
				t.preCount[nextTask] -= 1
			}
			batch = append(batch, tName)
			t.taskSet.Del(tName)
		}

		if t.taskSet.Len() > 0 && len(batch) > 0 {
			return nil, errors.New("task graph has a loop")
		}

		result = append(result, batch)
	}

	return
}

func NewTaskDep() *TaskDep {
	return &TaskDep{
		taskSet:   NewTaskNameSet(),
		taskEdges: map[flow.TName][]flow.TName{},
		preCount:  map[flow.TName]int{},
	}
}
