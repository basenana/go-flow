package eventbus

import "sync"

var (
	eb *eventbus
)

func init() {
	eb = &eventbus{
		stone:    map[string]*listener{},
		registry: map[Topic]*listener{},
	}
}

type eventbus struct {
	stone    map[string]*listener
	registry map[Topic]*listener
	mux      sync.RWMutex
}

func (b *eventbus) register(t Topic, l *listener) {
	l.Topic = t

	b.mux.Lock()
	b.stone[l.ID] = l
	firstL := b.registry[t]
	l.next = firstL
	b.registry[t] = l
	b.mux.Unlock()
}

func (b *eventbus) unregister(ID string) {
	b.mux.Lock()
	defer b.mux.Unlock()

	l, ok := b.stone[ID]
	if !ok {
		return
	}
	delete(b.stone, ID)

	lastL := b.registry[l.Topic]
	if lastL.ID == ID {
		if lastL.next == nil {
			delete(b.registry, l.Topic)
		} else {
			b.registry[l.Topic] = lastL.next
		}
		return
	}

	crtL := lastL.next
	for crtL != nil {
		switch {
		case crtL.ID == ID:
			lastL.next = crtL.next
			return
		default:
			lastL = lastL.next
			crtL = lastL.next
		}
	}
}

func (b *eventbus) publish(t Topic, args ...interface{}) {
	b.mux.RLock()
	defer b.mux.RUnlock()

	l := b.registry[t]
	for l != nil {
		l.do(args...)
		l = l.next
	}
}

func Register(t Topic, l *listener) {
	eb.register(t, l)
}

func Unregister(ID string) {
	eb.unregister(ID)
}

func Publish(t Topic, args ...interface{}) {
	eb.publish(t, args...)
}
