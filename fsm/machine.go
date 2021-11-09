package fsm

import (
	"go-flow/eventbus"
	"go-flow/log"
)

type FSM struct {
	logger log.Logger
}

func (m *FSM) From(statues []string) *FSM {
	return m
}

func (m *FSM) To(statue string) *FSM {
	return m
}

func (m *FSM) When(event eventbus.Topic) *FSM {
	return m
}

func (m *FSM) Do(handler Handler) *FSM {
	return m
}

func (m *FSM) Close() error {
	return nil
}

func New(option Option) *FSM {
	return &FSM{}
}
