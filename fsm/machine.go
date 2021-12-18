package fsm

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/zwwhdls/go-flow/eventbus"
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
	name  string
	obj   Stateful
	graph map[EventType]*edge

	crtBuilder *edgeBuilder
	mux        sync.Mutex
	logger     log.Logger

	eventChs []chan Event
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

func (m *FSM) RegisterStatusCh(ch chan Event) {
	m.eventChs = append(m.eventChs, ch)
}

func (m *FSM) EventCh() chan Event {
	mergeCh := make(chan Event, 8)
	m.eventChs = append(m.eventChs, mergeCh)
	return mergeCh
}

func (m *FSM) Close() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	eventbus.Unregister(m.name)
	for _, ch := range m.eventChs {
		close(ch)
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
		m.graph[builder.when] = newEdge
	}
}

func (m *FSM) eventHandler(obj interface{}, args ...interface{}) (err error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	event, ok := obj.(Event)
	if !ok {
		err = fmt.Errorf("not got event object")
		return
	}

	head := m.graph[event.Type]
	if head == nil {
		return nil
	}

	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("event %s handle panic: %v", event, panicErr)
		}
	}()

	for head != nil {
		if m.obj.GetStatus() == head.from {
			m.obj.SetStatus(head.to)
			if event.Message != "" {
				m.obj.SetMessage(event.Message)
			}

			for _, ch := range m.eventChs {
				select {
				case ch <- Event{
					Type:    event.Type,
					Status:  m.obj.GetStatus(),
					Message: event.Message,
					Obj:     event.Obj,
				}:
				default:
				}
			}
			return head.do(event)
		}
		head = head.next
	}
	return nil
}

func New(option Option) *FSM {
	f := &FSM{
		name:     fmt.Sprintf("%s.%s", option.Name, uuid.New().String()),
		obj:      option.Obj,
		graph:    map[EventType]*edge{},
		logger:   option.Logger,
		eventChs: make([]chan Event, 0),
	}
	eventbus.Register(option.Topic, eventbus.NewSimpleListener(f.name, f.eventHandler))
	return f
}
