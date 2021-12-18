package flow

import (
	"context"
)

type Context struct {
	context.Context
	FlowId   FID
	MaxRetry *int
}

func (c *Context) Succeed() {
}

func (c *Context) Fail(MaxRetry int) {
	c.MaxRetry = &MaxRetry
}
