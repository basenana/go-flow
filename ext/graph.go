package ext

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
	"github.com/zwwhdls/go-flow/storage"
)

type GraphFlow struct {
	*basic
	tasks   map[flow.TName]*GraphTask
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
	ctx.Succeed()
	return nil
}

func (g GraphFlow) Teardown(ctx *flow.Context) {
	return
}

func (g GraphFlow) taskDoFunc(ctx *flow.Context, appId string) error {
	// define your task func here
	ctx.Logger.Infof("app %s task done", appId)
	ctx.Succeed()
	return nil
}

func (g GraphFlow) taskTearDownFunc(ctx *flow.Context, appId string) {
	// define your task func here
	ctx.Logger.Infof("app %s task teardown", appId)
	ctx.Succeed()
}

func (g GraphFlow) taskSetUpFunc(ctx *flow.Context, appId string) error {
	// define your task func here
	ctx.Logger.Infof("app %s task setup", appId)
	ctx.Succeed()
	return nil
}

func (g *GraphFlow) NextBatch(ctx *flow.Context) ([]flow.Task, error) {
	if len(g.batches) == 0 {
		return nil, nil
	}

	crtBatch := g.batches[0]
	g.batches = g.batches[1:]

	tasks := make([]flow.Task, len(crtBatch))
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

func NewGraphFlow(s storage.Interface, tasks []*GraphTask, dep *TaskDep, policy flow.ControlPolicy) (flow.Flow, error) {
	if dep == nil || dep.taskEdges == nil {
		dep = NewTaskDep()
	}

	batches, err := dep.Lists()
	if err != nil {
		return nil, err
	}

	taskMap := make(map[flow.TName]*GraphTask)
	for i, t := range tasks {
		taskMap[t.TaskName] = tasks[i]
	}

	f := &GraphFlow{
		basic: &basic{
			id:     flow.FID(fmt.Sprintf("graph-flow-%s", uuid.New().String())),
			policy: policy,
		},
		tasks:   taskMap,
		batches: batches,
	}
	return f, s.SaveFlow(f)
}

type GraphTask struct {
	TaskName flow.TName
	AppId    string
	Message  string
	status   fsm.Status

	setupFunc    func(ctx *flow.Context) error
	doFunc       func(ctx *flow.Context) error
	teardownFunc func(ctx *flow.Context)
}

func (g GraphTask) GetStatus() fsm.Status {
	return g.status
}

func (g *GraphTask) SetStatus(status fsm.Status) {
	g.status = status
}

func (g GraphTask) GetMessage() string {
	return g.Message
}

func (g *GraphTask) SetMessage(msg string) {
	g.Message = msg
}

func (g GraphTask) Name() flow.TName {
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
		TaskName: flow.TName(name),
		AppId:    appId,
		Message:  msg,
		status:   flow.CreatingStatus,
	}
}

type TaskDep struct {
	taskSet   TaskNameSet
	taskEdges map[flow.TName][]flow.TName
	preCount  map[flow.TName]int
}

func (t *TaskDep) RunOrder(firstTask flow.TName, tasks []flow.TName) {
	t.taskSet.Insert(firstTask)
	for _, task := range tasks {
		t.taskSet.Insert(task)

		edges, ok := t.taskEdges[firstTask]
		if !ok {
			edges = make([]flow.TName, 0)
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

func (t *TaskDep) Lists() (result [][]flow.TName, err error) {

	for t.taskSet.Len() != 0 {
		batch := make([]flow.TName, 0)
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
		taskEdges: map[flow.TName][]flow.TName{},
		preCount:  map[flow.TName]int{},
	}
}
