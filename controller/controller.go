package controller

import (
	"context"
	"fmt"
	"github.com/zwwhdls/go-flow/ext"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
	"github.com/zwwhdls/go-flow/log"
	"github.com/zwwhdls/go-flow/storage"
	"reflect"
)

type FlowController struct {
	flows   map[flow.FID]*runner
	storage storage.Interface
	logger  log.Logger
}

func (c *FlowController) NewFlow(builder ext.FlowBuilder) (flow.Flow, error) {
	f := builder.Build()
	if reflect.ValueOf(f).Type().Kind() != reflect.Ptr {
		return nil, fmt.Errorf("flow object must be ptr")
	}

	c.logger.Infof("build flow %s", f.ID())
	f.SetStatus(flow.CreatingStatus)
	c.flows[f.ID()] = &runner{
		Flow:    f,
		stopCh:  make(chan struct{}),
		storage: c.storage,
		logger:  c.logger.With(fmt.Sprintf("flow.%s", f.ID())),
	}
	return f, nil
}

func (c *FlowController) TriggerFlow(ctx context.Context, flowId flow.FID) error {
	r, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}

	c.logger.Infof("trigger flow %s", flowId)
	return r.Start(&flow.Context{
		Context: ctx,
		Logger:  c.logger.With(fmt.Sprintf("flow.%s", flowId)),
		FlowId:  flowId,
	})
}

func (c *FlowController) PauseFlow(flowId flow.FID) error {
	r, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	if r.GetStatus() == flow.RunningStatus {
		c.logger.Infof("pause flow %s", flowId)
		return r.Pause(fsm.Event{
			Type:   flow.ExecutePauseEvent,
			Status: r.GetStatus(),
			Obj:    r.Flow,
		})
	}
	return fmt.Errorf("flow current is %s, can not pause", r.GetStatus())
}

func (c *FlowController) CancelFlow(flowId flow.FID) error {
	r, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	switch r.GetStatus() {
	case flow.RunningStatus, flow.PausedStatus:
		c.logger.Infof("cancel flow %s", flowId)
		return r.Cancel(fsm.Event{
			Type:    flow.ExecuteCancelEvent,
			Status:  r.GetStatus(),
			Message: "canceled",
			Obj:     r.Flow,
		})
	default:
		return fmt.Errorf("flow current is %s, can not cancel", r.GetStatus())
	}
}

func (c *FlowController) ResumeFlow(flowId flow.FID) error {
	r, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	switch r.GetStatus() {
	case flow.PausedStatus:
		c.logger.Infof("resume flow %s", flowId)
		return r.Resume(fsm.Event{
			Type:   flow.ExecuteResumeEvent,
			Status: r.GetStatus(),
			Obj:    r.Flow,
		})
	default:
		return fmt.Errorf("flow current is %s, can not resume", r.GetStatus())
	}
}

func (c *FlowController) CleanFlow(flowId flow.FID) error {
	r, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	switch r.GetStatus() {
	case flow.SucceedStatus, flow.FailedStatus, flow.CanceledStatus:
		c.logger.Infof("clean flow %s", flowId)
		delete(c.flows, flowId)
	default:
		return fmt.Errorf("flow current is %s, can not clean", r.GetStatus())
	}
	return nil
}

func NewFlowController(opt Option) (*FlowController, error) {
	return &FlowController{
		flows:   make(map[flow.FID]*runner),
		storage: opt.Storage,
		logger:  log.NewLogger("go-flow"),
	}, nil
}
