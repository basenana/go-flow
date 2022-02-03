package fsm

import (
	"fmt"
	"github.com/zwwhdls/go-flow/log"
	"sync"
)

type edge struct {
	from Status
	to   Status
	when EventType
	do   Handler
	next *edge
}

type edgeBuilder struct {
	from []Status
	to   Status
	when EventType
	do   Handler
}

type FSM struct {
	obj   Stateful
	graph map[EventType]*edge

	crtBuilder *edgeBuilder
	mux        sync.Mutex
	logger     log.Logger
}

func (m *FSM) From(statues []Status) *FSM {
	m.buildWarp(func(builder *edgeBuilder) {
		builder.from = statues
	})
	return m
}

func (m *FSM) To(status Status) *FSM {
	m.buildWarp(func(builder *edgeBuilder) {
		builder.to = status
	})
	return m
}

func (m *FSM) When(event EventType) *FSM {
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

func (m *FSM) Event(event Event) error {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.logger.Debugf("handler fsm event: %s", event.Type)
	head := m.graph[event.Type]
	if head == nil {
		return nil
	}

	defer func() {
		if panicErr := recover(); panicErr != nil {
			m.logger.Errorf("event %s handle panic: %v", event, panicErr)
		}
	}()

	for head != nil {
		if m.obj.GetStatus() == head.from {
			m.logger.Infof("change obj status from %s to %s with event: %s", head.from, head.to, event.Type)
			m.obj.SetStatus(head.to)
			if event.Message != "" {
				m.obj.SetMessage(event.Message)
			}

			if head.do != nil {
				if handleErr := head.do(event); handleErr != nil {
					m.logger.Errorf("event %s handle failed: %s", event.Type, handleErr.Error())
					return handleErr
				}
			}
			return nil
		}
		head = head.next
	}
	return fmt.Errorf("get event %s and current status is %s, no change path matched", event.Type, m.obj.GetStatus())
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
		m.graph[builder.when] = newEdge
	}
}

func New(option Option) *FSM {
	f := &FSM{
		obj:    option.Obj,
		graph:  map[EventType]*edge{},
		logger: option.Logger,
	}
	return f
}
