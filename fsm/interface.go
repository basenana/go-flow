package fsm

type Stateful interface {
	GetStatus() string
	SetStatus(string) error
}

type Handler func() error
