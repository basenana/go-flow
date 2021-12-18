package fsm

type Stateful interface {
	GetStatus() Status
	SetStatus(status Status)
	GetMessage() string
	SetMessage(msg string)
}

type Handler func(obj Event) error
