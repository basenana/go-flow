package flow

import (
	"context"
	"github.com/zwwhdls/go-flow/eventbus"
	"sync"
)

type Context struct {
	sync.Mutex
	context.Context
	FlowId   FID
	TaskName TName
	MaxRetry int
	Retry    int
}

func (c *Context) Succeed() {
	eventbus.Publish(GetTaskTopic(TaskExecuteSucceedEventTopicTpl, c.FlowId, c.TaskName))
}

func (c *Context) Fail(MaxRetry int) {
	c.MaxRetry = MaxRetry
	eventbus.Publish(GetTaskTopic(TaskExecuteFailedEventTopicTpl, c.FlowId, c.TaskName))
}
