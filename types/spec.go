/*
   Copyright 2023 Go-Flow Authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package types

const (
	InitializingStatus = "initializing"
	RunningStatus      = "running"
	SucceedStatus      = "succeed"
	FailedStatus       = "failed"
	ErrorStatus        = "error"
	PausedStatus       = "paused"
	CanceledStatus     = "canceled"

	TriggerEvent       = "flow.execute.trigger"
	ExecuteFinishEvent = "flow.execute.finish"
	ExecuteFailedEvent = "flow.execute.failed"
	ExecuteErrorEvent  = "flow.execute.error"
	ExecutePauseEvent  = "flow.execute.pause"
	ExecuteResumeEvent = "flow.execute.resume"
	ExecuteCancelEvent = "flow.execute.cancel"

	PolicyFastFailed = "fastFailed"
	PolicyPaused     = "paused"
	PolicyContinue   = "continue"
)

type Spec struct {
	Type       string            `json:"type"`
	Http       *Http             `json:"http,omitempty"`
	Translate  *Translate        `json:"translate,omitempty"`
	Script     *Script           `json:"script,omitempty"`
	Parameters map[string]string `json:"parameters"`
}

type Script struct {
	Content string            `json:"content"`
	Command []string          `json:"command"`
	Env     map[string]string `json:"env"`
}

type Http struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	ContentType string            `json:"content_type"`
	Body        string            `json:"body"`
	Query       map[string]string `json:"query"`
	Headers     map[string]string `json:"headers"`
	Insecure    bool              `json:"insecure"`
}

type Translate struct {
	Encoder *Encoder `json:"encoder,omitempty"`
	Decoder *Decoder `json:"decoder,omitempty"`
}

const (
	EncodeTypeGoTpl = "gotemplate"
)

type Encoder struct {
	EncodeType string `json:"encode_type"`
	Template   string `json:"template"`
}

const (
	DecodeTypeJsonPath = "jsonpath"
	DecodeTypeRegular  = "regular"
)

type Decoder struct {
	DecodeType string `json:"decode_type"`
	InputTask  string `json:"input_task"`
	Pattern    string `json:"pattern"`
}

type NextTask struct {
	OnSucceed string `json:"on_succeed,omitempty"`
	OnFailed  string `json:"on_failed,omitempty"`
}

type ControlPolicy struct {
	FailedPolicy string
}

func IsFinishedStatus(sts string) bool {
	switch sts {
	case SucceedStatus, FailedStatus, CanceledStatus, ErrorStatus:
		return true
	default:
		return false
	}
}
