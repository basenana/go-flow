package fsm

import "go-flow/log"

type Option struct {
	Name   string
	Obj    Stateful
	Logger log.Logger
}
