package flow

import (
	"context"
	"github.com/zwwhdls/go-flow/log"
)

type Context struct {
	context.Context
	log.Logger

	FlowId    FID
	TaskName  TName
	Message   string
	MaxRetry  int
	IsSucceed bool
}

func (c *Context) Succeed() {
	c.IsSucceed = true
}

func (c *Context) Fail(message string, maxRetry int) {
	c.Message = message
	c.MaxRetry = maxRetry
}
