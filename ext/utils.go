package ext

import (
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
)

type basic struct {
	id      flow.FID
	status  fsm.Status
	message string
	policy  flow.ControlPolicy
}

func (b basic) ID() flow.FID {
	return b.id
}

func (b basic) GetStatus() fsm.Status {
	return b.status
}

func (b *basic) SetStatus(status fsm.Status) {
	b.status = status
}

func (b basic) GetMessage() string {
	return b.message
}

func (b *basic) SetMessage(msg string) {
	b.message = msg
}

func (b basic) GetControlPolicy() flow.ControlPolicy {
	return b.policy
}

type TaskNameSet map[flow.TName]struct{}

func (s TaskNameSet) Insert(t flow.TName) {
	s[t] = struct{}{}
}
func (s TaskNameSet) Has(task flow.TName) bool {
	_, ok := s[task]
	return ok
}

func (s TaskNameSet) Del(t flow.TName) {
	if _, ok := s[t]; ok {
		delete(s, t)
	}
}

func (s TaskNameSet) List() (result []flow.TName) {
	for k := range s {
		result = append(result, k)
	}
	return
}

func (s TaskNameSet) Len() int {
	return len(s)
}

func NewTaskNameSet() TaskNameSet {
	return map[flow.TName]struct{}{}
}
