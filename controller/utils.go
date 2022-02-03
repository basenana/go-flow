package controller

import (
	"bytes"
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
