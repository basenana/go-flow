package fsm

import (
	"go-flow/eventbus"
	"go-flow/log"
)

type machine struct {
	logger log.Logger
}

func (m *machine) From(statues []string) *machine {
	return m
}

func (m *machine) To(statue string) *machine {
	return m
}

func (m *machine) When(event eventbus.Topic) *machine {
	return m
}

func (m *machine) Do(handler Handler) *machine {
	return m
}

func (m *machine) Close() error {
	return nil
}

func New(option Option) *machine {
	return &machine{}
}
