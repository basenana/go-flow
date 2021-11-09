package controller

import (
	"github.com/zwwhdls/go-flow/log"
	"github.com/zwwhdls/go-flow/storage"
)

type Option struct {
	Storage storage.Interface
	Logger  log.Logger
}
