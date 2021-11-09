package fsm

import "github.com/zwwhdls/go-flow/log"

type Option struct {
	Name   string
	Obj    Stateful
	Logger log.Logger
}
