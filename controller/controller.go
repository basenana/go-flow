package controller

import (
	"context"
	"fmt"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
	"github.com/zwwhdls/go-flow/log"
	"github.com/zwwhdls/go-flow/storage"
)

type FlowController struct {
	flows   map[flow.FID]*runner
	storage storage.Interface
	logger  log.Logger
}

func (c *FlowController) TriggerFlow(ctx context.Context, flowId flow.FID) error {
	f, err := c.storage.GetFlow(flowId)
	if err != nil {
		return err
	}

	r := &runner{
		Flow:    f,
		stopCh:  make(chan struct{}),
		storage: c.storage,
		logger:  c.logger.With(fmt.Sprintf("flow.%s", f.ID())),
	}
	c.flows[f.ID()] = r

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
