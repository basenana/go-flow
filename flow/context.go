package flow

import (
	"context"
	"github.com/zwwhdls/go-flow/eventbus"
)

type FlowContext struct {
	context.Context
	FlowId FID
}

type TaskContext struct {
	context.Context
	Flow     FlowContext
	TaskName TName
}

func (tc *TaskContext) Succeed() {
	eventbus.Publish(GetTaskTopic(TaskExecuteSucceedEventTopicTpl, tc.Flow.FlowId, tc.TaskName))
}

func (tc *TaskContext) Fail() {
	eventbus.Publish(GetTaskTopic(TaskExecuteFailedEventTopicTpl, tc.Flow.FlowId, tc.TaskName))
}
