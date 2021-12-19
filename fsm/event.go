package fsm

type EventType string

type Status string

type Event struct {
	Type    EventType
	Status  Status
	Message string
	Obj     Stateful
}
