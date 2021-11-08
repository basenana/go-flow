package controller

import (
	"go-flow/log"
	"go-flow/storage"
)

type Option struct {
	Storage storage.Interface
	Logger  log.Logger
}
