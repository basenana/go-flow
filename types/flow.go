package types

type Flow interface {
	GetID() string
	GetStatus() string
	SetStatus(string)
	GetMessage() string
	SetMessage(string)
	GetExecutor() string
	GetControlPolicy() ControlPolicy
	Tasks() []Task
}

type FlowSpec struct {
	ID            string        `json:"id"`
	Describe      string        `json:"describe"`
	Executor      string        `json:"executor"`
	Scheduler     string        `json:"scheduler"`
	Status        string        `json:"status"`
	Message       string        `json:"message"`
	ControlPolicy ControlPolicy `json:"control_policy"`
	Tasks         []TaskSpec    `json:"tasks"`
	OnFailure     []TaskSpec    `json:"on_failure"`
}

func (f *FlowSpec) GetStatus() string {
	return f.Status
}

func (f *FlowSpec) SetStatus(status string) {
	f.Status = status
}

func (f *FlowSpec) GetMessage() string {
	return f.Message
}

func (f *FlowSpec) SetMessage(msg string) {
	f.Message = msg
}
