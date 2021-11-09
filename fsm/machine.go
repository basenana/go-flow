package fsm

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/zwwhdls/go-flow/eventbus"
	"github.com/zwwhdls/go-flow/log"
	"sync"
)

type edge struct {
	from string
	to   string
	when eventbus.Topic
	do   Handler
	next *edge
}

type edgeBuilder struct {
	from []string
	to   string
	when eventbus.Topic
	do   Handler
}

type FSM struct {
	name      string
	obj       Stateful
	graph     map[eventbus.Topic]*edge
	listeners []string

	crtBuilder *edgeBuilder
	mux        sync.Mutex
	logger     log.Logger
}

func (m *FSM) From(statues []string) *FSM {
	m.buildWarp(func(builder *edgeBuilder) {
		builder.from = statues
	})
	return m
}

func (m *FSM) To(status string) *FSM {
	m.buildWarp(func(builder *edgeBuilder) {
		builder.to = status
	})
	return m
}

func (m *FSM) When(event eventbus.Topic) *FSM {
	m.buildWarp(func(builder *edgeBuilder) {
		builder.when = event
	})
	return m
}

func (m *FSM) Do(handler Handler) *FSM {
	m.buildWarp(func(builder *edgeBuilder) {
		builder.do = handler
	})
	return m
}

func (m *FSM) Close() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	for _, lID := range m.listeners {
		eventbus.Unregister(lID)
	}
	return nil
}

func (m *FSM) buildWarp(f func(builder *edgeBuilder)) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if m.crtBuilder == nil {
		m.crtBuilder = &edgeBuilder{}
	}
	f(m.crtBuilder)

	if len(m.crtBuilder.from) == 0 {
		return
	}

	if m.crtBuilder.to == "" {
		return
	}

	if m.crtBuilder.when == "" {
		return
	}

	if m.crtBuilder.do == nil {
		return
	}

	builder := m.crtBuilder
	m.crtBuilder = nil

	head := m.graph[builder.when]

	for _, from := range builder.from {
		newEdge := &edge{
			from: from,
			to:   builder.to,
			when: builder.when,
			do:   builder.do,
			next: head,
		}
		head = newEdge

		lID := fmt.Sprintf("fsm.%s.%s", m.name, uuid.New().String())
		eventbus.Register(builder.when, eventbus.NewBlockListener(lID, func(args ...interface{}) error {
			if m.obj.GetStatus() != newEdge.from {
				return nil
			}
			if err := m.obj.SetStatus(newEdge.to); err != nil {
				return err
			}
			return builder.do(args...)
		}))
		m.listeners = append(m.listeners, lID)
	}
}

func New(option Option) *FSM {
	return &FSM{}
}
