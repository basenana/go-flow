package controller

import (
	"github.com/zwwhdls/go-flow/ext"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
)

type flowWarp struct {
	flow.Flow
	ctx          *flow.Context
	machine      *fsm.FSM
	builder      ext.FlowBuilder
	currentTasks []flow.Task
	flowCh       chan string
}

type taskWarp struct {
	flow.Task
	ctx     *flow.Context
	fid     flow.FID
	machine *fsm.FSM
}
