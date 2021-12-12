package controller

import (
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
	"github.com/zwwhdls/go-flow/plugin"
)

type flowWarp struct {
	flow.Flow
	ctx          *flow.Context
	machine      *fsm.FSM
	builder      plugin.FlowBuilder
	currentTasks []flow.Task
	flowCh       chan string
}

type taskWarp struct {
	flow.Task
	ctx     *flow.Context
	fid     flow.FID
	machine *fsm.FSM
}
