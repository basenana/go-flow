package controller

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/zwwhdls/go-flow/ext"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
	"math/rand"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var controller *FlowController

func init() {
	controller = NewFlowController(Option{})
}

func TestController(t *testing.T) {
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
		return f.chaos.FlowSetupErr
	}
	return nil
}

func (f ChaosFlow) Teardown(ctx *flow.Context) {
	f.doIdle()
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
		crtCtn := f.taskCount
		batchTasks[i] = ext.BuildSampleTask(fmt.Sprintf("task-%d", crtCtn), ext.TaskSpec{
			Setup: func(ctx *flow.Context) error {
				f.doIdle()
				return f.chaos.TaskSetupErr
			},
			Do: func(ctx *flow.Context) error {
				f.doIdle()
				if f.chaos.FailTaskIdx > 0 && f.chaos.FailTaskIdx == crtCtn {
					return f.chaos.TaskErr
				}
				return nil
			},
			Teardown: func(ctx *flow.Context) {
				f.doIdle()
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

type ChaosFlowBuilder struct {
	BatchNum int
	Parallel int
	Policy   flow.ControlPolicy
	Chaos    ChaosPolicy
}

var _ ext.FlowBuilder = ChaosFlowBuilder{}

func (f ChaosFlowBuilder) Build() flow.Flow {
	return &ChaosFlow{
		id:        flow.FID(fmt.Sprintf("chaos-flow-%s", uuid.New().String())),
		batchLeft: f.BatchNum,
		parallel:  f.Parallel,
		status:    flow.CreatingStatus,
		policy:    f.Policy,
		chaos:     f.Chaos,
	}
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
