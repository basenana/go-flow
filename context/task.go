package context

import "context"

type TaskContext struct {
	context.Context
	Flow FlowContext
}
