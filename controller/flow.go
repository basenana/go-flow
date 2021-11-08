package controller

import (
	"go-flow/flow"
	"go-flow/plugin"
)

func (c *FlowController) NewFlow(builder plugin.FlowBuilder) flow.Flow {
	return nil
}

func (c *FlowController) TriggerFlow(flowId flow.FID) error {
	return nil
}

func (c *FlowController) PauseFlow(flowId flow.FID) error {
	return nil
}

func (c *FlowController) CancelFlow(flowId flow.FID) error {
	return nil
}
