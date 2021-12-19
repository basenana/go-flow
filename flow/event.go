package flow

import (
	"fmt"
	"github.com/zwwhdls/go-flow/eventbus"
	"github.com/zwwhdls/go-flow/fsm"
	"strings"
)

const flowTopic = "event.flow"

func EventTopic(fid FID) eventbus.Topic {
	return eventbus.Topic(fmt.Sprintf("%s.%s", flowTopic, fid))
}

func IsTaskEvent(eventType fsm.EventType) bool {
	return strings.HasPrefix(string(eventType), "task.")
}
