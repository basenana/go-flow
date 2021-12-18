package controller

import (
	"context"
	"fmt"
	"github.com/zwwhdls/go-flow/eventbus"
	"github.com/zwwhdls/go-flow/ext"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
	"github.com/zwwhdls/go-flow/log"
	"github.com/zwwhdls/go-flow/storage"
	"reflect"
)

type FlowController struct {
	flows   map[flow.FID]*runner
	Storage storage.Interface
	Logger  log.Logger
}

func (c *FlowController) NewFlow(builder ext.FlowBuilder) (flow.Flow, error) {
	f := builder.Build()
	if reflect.ValueOf(f).Type().Kind() != reflect.Ptr {
		return nil, fmt.Errorf("flow object must be ptr")
	}

	c.Logger.Infof("Build flow %s", f.ID())
	f.SetStatus(flow.CreatingStatus)
	c.flows[f.ID()] = &runner{
		Flow:   f,
		stopCh: make(chan struct{}),
		logger: c.Logger,
	}
	return f, nil
}

func (c *FlowController) TriggerFlow(ctx context.Context, flowId flow.FID) error {
	c.Logger.Infof("Trigger flow %s", flowId)
	r, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}

	return r.start(flow.Context{
		Context: ctx,
		FlowId:  flowId,
	})
}

func (c *FlowController) PauseFlow(flowId flow.FID) error {
	c.Logger.Infof("Pause flow %s", flowId)
	r, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	if r.GetStatus() == flow.RunningStatus {
		eventbus.Publish(flow.EventTopic(flowId), fsm.Event{
			Type:   flow.ExecutePauseEvent,
			Status: r.GetStatus(),
			Obj:    r.Flow,
		})
	}
	return fmt.Errorf("flow current is %s, can not pause", r.GetStatus())
}

func (c *FlowController) CancelFlow(flowId flow.FID) error {
	c.Logger.Infof("Cancel flow %s", flowId)
	r, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	switch r.GetStatus() {
	case flow.RunningStatus, flow.PausedStatus:
		eventbus.Publish(flow.EventTopic(flowId), fsm.Event{
			Type:   flow.ExecuteCancelEvent,
			Status: r.GetStatus(),
			Obj:    r.Flow,
		})
		return nil
	default:
		return fmt.Errorf("flow current is %s, can not cancel", r.GetStatus())
	}
}

func (c *FlowController) ResumeFlow(flowId flow.FID) error {
	c.Logger.Infof("Resume flow %s", flowId)
	r, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	switch r.GetStatus() {
	case flow.PausedStatus:
		eventbus.Publish(flow.EventTopic(flowId), fsm.Event{
			Type:   flow.ExecuteResumeEvent,
			Status: r.GetStatus(),
			Obj:    r.Flow,
		})
		return nil
	default:
		return fmt.Errorf("flow current is %s, can not resume", r.GetStatus())
	}
}

func NewFlowController(opt Option) *FlowController {
	return &FlowController{
		Storage: opt.Storage,
		Logger:  opt.Logger,
		flows:   make(map[flow.FID]*runner),
	}
}
