package controller

import (
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/log"
	"github.com/zwwhdls/go-flow/storage"
)

type FlowController struct {
	flows   map[flow.FID]*flowWarp
	Storage storage.Interface
	Logger  log.Logger
}

func NewFlowController(opt Option) *FlowController {
	return &FlowController{
		Storage: opt.Storage,
		Logger:  opt.Logger,
		flows:   make(map[flow.FID]*flowWarp),
	}
}
