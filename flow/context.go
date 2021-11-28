package flow

import (
	"context"
	"github.com/zwwhdls/go-flow/eventbus"
	"sync"
)

type FlowContext struct {
	sync.Mutex
	context.Context
	FlowId  FID
	Operate FlowOperate
}

type TaskContext struct {
	sync.Mutex
	context.Context
	Flow     *FlowContext
	TaskName TName
	Operate  TaskOperate
	MaxRetry int
	Retry    int
}

type TaskOperate string

var (
	TaskInitOperate     TaskOperate = "task.init"
	TaskExecuteOperate  TaskOperate = "task.execute"
	TaskTeardownOperate TaskOperate = "task.teardown"
)

type FlowOperate string

var (
	FlowInitOperate     FlowOperate = "flow.init"
	FlowExecuteOperate  FlowOperate = "flow.execute"
	FlowTeardownOperate FlowOperate = "flow.teardown"
)

func (tc *TaskContext) Succeed() {
	eventbus.Publish(GetTaskTopic(TaskExecuteSucceedEventTopicTpl, tc.Flow.FlowId, tc.TaskName))
}

func (tc *TaskContext) Fail(MaxRetry int) {
	tc.MaxRetry = MaxRetry
	eventbus.Publish(GetTaskTopic(TaskExecuteFailedEventTopicTpl, tc.Flow.FlowId, tc.TaskName))
}
