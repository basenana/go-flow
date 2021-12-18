package flow

type FailedPolicy string

const (
	PolicyFastFailed = "fastFailed"
	PolicyPaused     = "paused"
)

type ControlPolicy struct {
	FailedPolicy FailedPolicy
}
