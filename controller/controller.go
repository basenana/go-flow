package controller

import (
	"go-flow/log"
	"go-flow/storage"
)

type FlowController struct {
	Storage storage.Interface
	Logger  log.Logger
}

func NewFlowController(opt Option) *FlowController {
	return &FlowController{
		Storage: opt.Storage,
		Logger:  opt.Logger,
	}
}
