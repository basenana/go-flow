package fsm

import (
	"github.com/zwwhdls/go-flow/eventbus"
	"github.com/zwwhdls/go-flow/log"
)

type Option struct {
	Name   string
	Obj    Stateful
	Topic  eventbus.Topic
	Filter func(event Event) bool
	Logger log.Logger
}
