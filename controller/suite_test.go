package controller

import (
	"fmt"
	"github.com/onsi/ginkgo/config"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
	"github.com/zwwhdls/go-flow/storage"
	"math/rand"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	controller  *FlowController
	flowStorage storage.Interface
)

func init() {
	flowStorage = storage.NewInMemoryStorage()
	controller, _ = NewFlowController(Option{Storage: flowStorage})
}

func TestController(t *testing.T) {
	config.DefaultReporterConfig.SlowSpecThreshold = time.Hour.Seconds()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

type ChaosFlow struct {
	id        flow.FID
	batchLeft int
	parallel  int
	taskCount int
	status    fsm.Status
	message   string
	policy    flow.ControlPolicy
	chaos     ChaosPolicy
}

var _ flow.Flow = &ChaosFlow{}

func (f ChaosFlow) Type() flow.FType {
	return "chaos"
}

func (f ChaosFlow) GetStatus() fsm.Status {
	return f.status
}

func (f *ChaosFlow) SetStatus(status fsm.Status) {
	f.status = status
}

func (f ChaosFlow) GetMessage() string {
	return f.message
}

func (f *ChaosFlow) SetMessage(msg string) {
	f.message = msg
}

func (f ChaosFlow) ID() flow.FID {
	return f.id
}

func (f ChaosFlow) GetHooks() flow.Hooks {
	return map[flow.HookType]flow.Hook{}
}

func (f ChaosFlow) Setup(ctx *flow.Context) error {
	f.doIdle()
	if f.chaos.FlowSetupErr != nil {
		ctx.Fail(f.chaos.FlowSetupErr.Error(), 3)
		return f.chaos.FlowSetupErr
	}
	ctx.Succeed()
	return nil
}

func (f ChaosFlow) Teardown(ctx *flow.Context) {
	f.doIdle()
	ctx.Succeed()
	return
}

func (f *ChaosFlow) NextBatch(ctx *flow.Context) ([]flow.Task, error) {
	if f.chaos.GetBatchErr != nil {
		return nil, f.chaos.GetBatchErr
	}

	if f.batchLeft == 0 {
		return nil, nil
	}
	f.batchLeft -= 1

	batchTasks := make([]flow.Task, f.parallel)
	for i := 0; i < f.parallel; i++ {
		crtCtn := f.taskCount + 1
		batchTasks[i] = BuildSampleTask(fmt.Sprintf("task-%d", crtCtn), TaskSpec{
			Setup: func(ctx *flow.Context) error {
				f.doIdle()
				if f.chaos.FailTaskIdx == crtCtn && f.chaos.TaskSetupErr != nil {
					ctx.Fail(f.chaos.TaskSetupErr.Error(), 3)
					return f.chaos.TaskSetupErr
				}
				ctx.Succeed()
				return nil
			},
			Do: func(ctx *flow.Context) error {
				f.doIdle()
				if f.chaos.FailTaskIdx > 0 && f.chaos.FailTaskIdx == crtCtn {
					ctx.Fail(f.chaos.TaskErr.Error(), 3)
					return f.chaos.TaskErr
				}
				ctx.Succeed()
				return nil
			},
			Teardown: func(ctx *flow.Context) {
				f.doIdle()
				ctx.Succeed()
			},
		})
		f.taskCount += 1
	}

	return batchTasks, nil
}

func (f ChaosFlow) GetControlPolicy() flow.ControlPolicy {
	return f.policy
}

func (f ChaosFlow) doIdle() {
	sleepTime := rand.Int()%5 + 1
	time.Sleep(time.Duration(sleepTime) * time.Second)
}

type ChaosPolicy struct {
	FailTaskIdx  int
	TaskErr      error
	FlowSetupErr error
	TaskSetupErr error
	GetBatchErr  error
}

func status2Bytes(status fsm.Status) []byte {
	return []byte(status)
}

type Task struct {
	name     flow.TName
	taskType string
	appId    string
	status   fsm.Status
	message  string

	setupFunc    func(ctx *flow.Context) error
	doFunc       func(ctx *flow.Context) error
	teardownFunc func(ctx *flow.Context)
}

var _ flow.Task = &Task{}

func (t Task) GetStatus() fsm.Status {
	return t.status
}

func (t *Task) SetStatus(status fsm.Status) {
	t.status = status
}

func (t Task) GetMessage() string {
	return t.message
}

func (t *Task) SetMessage(msg string) {
	t.message = msg
}

func (t Task) Name() flow.TName {
	return t.name
}

func (t Task) Setup(ctx *flow.Context) error {
	return t.setupFunc(ctx)
}

func (t Task) Do(ctx *flow.Context) error {
	return t.doFunc(ctx)
}

func (t Task) Teardown(ctx *flow.Context) {
	t.teardownFunc(ctx)
}

func BuildSampleTask(uniqueName string, spec TaskSpec) flow.Task {
	return &Task{
		name:   flow.TName(uniqueName),
		status: flow.CreatingStatus,

		setupFunc:    spec.Setup,
		doFunc:       spec.Do,
		teardownFunc: spec.Teardown,
	}
}

type TaskSpec struct {
	Setup    func(ctx *flow.Context) error
	Do       func(ctx *flow.Context) error
	Teardown func(ctx *flow.Context)
}
