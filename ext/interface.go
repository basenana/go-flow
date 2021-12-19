package ext

import (
	"github.com/zwwhdls/go-flow/flow"
)

type FlowBuilder interface {
	Build() flow.Flow
}
