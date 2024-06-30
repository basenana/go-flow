package types

type Task interface {
	GetName() string
	GetStatue() string
	SetStatue(string)
	GetMessage() string
	SetMessage(string)
	Next() NextTask
}

type TaskSpec struct {
	Name          string   `json:"name"`
	Status        string   `json:"status"`
	Message       string   `json:"message"`
	OperatorSpec  Spec     `json:"operator_spec"`
	Next          NextTask `json:"next,omitempty"`
	RetryOnFailed int      `json:"retry_on_failed,omitempty"`
}

func (t *TaskSpec) GetStatus() string {
	return t.Status
}

func (t *TaskSpec) SetStatus(status string) {
	t.Status = status
}

func (t *TaskSpec) GetMessage() string {
	return t.Message
}

func (t *TaskSpec) SetMessage(msg string) {
	t.Message = msg
}
