package controller

import (
	"go-flow/flow"
	"go-flow/fsm"
	"go-flow/plugin"
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
