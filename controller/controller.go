package controller

import (
	"go-flow/flow"
	"go-flow/log"
	"go-flow/storage"
)

type FlowController struct {
	flows map[flow.FID]*flowWarp
	Storage storage.Interface
	Logger  log.Logger
}

func NewFlowController(opt Option) *FlowController {
	return &FlowController{
		Storage: opt.Storage,
		Logger:  opt.Logger,
	}
}
