package controller

import (
	"bytes"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/fsm"
)

type Errors []error

func (e Errors) Error() string {
	buf := bytes.Buffer{}
	for _, oneE := range e {
		buf.WriteString(oneE.Error())
		buf.WriteString(" ")
	}

	return buf.String()
}

func (e Errors) IsError() bool {
	return len(e) > 0
}

func NewErrors() Errors {
	return []error{}
}

func IsFinishedStatus(sts fsm.Status) bool {
	switch sts {
	case flow.SucceedStatus, flow.FailedStatus, flow.CanceledStatus:
		return true
	default:
		return false
	}
}
