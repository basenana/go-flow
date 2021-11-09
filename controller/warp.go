package controller

import (
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
	"github.com/zwwhdls/go-flow/plugin"
)

type flowWarp struct {
	flow.Flow
	machine      *fsm.FSM
	builder      plugin.FlowBuilder
	currentTasks []flow.Task
}

type taskWarp struct {
	flow.Task
	fid     flow.FID
	machine *fsm.FSM
}
