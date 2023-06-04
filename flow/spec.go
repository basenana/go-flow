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

package flow

import "github.com/basenana/go-flow/operator"

type Flow struct {
	ID            string        `json:"id"`
	Describe      string        `json:"describe"`
	Executor      string        `json:"executor"`
	Scheduler     string        `json:"scheduler"`
	Status        string        `json:"status"`
	Message       string        `json:"message"`
	ControlPolicy ControlPolicy `json:"control_policy"`
	Tasks         []Task        `json:"tasks"`
}

func (f *Flow) GetStatus() string {
	return f.Status
}

func (f *Flow) SetStatus(status string) {
	f.Status = status
}

func (f *Flow) GetMessage() string {
	return f.Message
}

func (f *Flow) SetMessage(msg string) {
	f.Message = msg
}

type Task struct {
	Name         string        `json:"name"`
	Status       string        `json:"status"`
	Message      string        `json:"message"`
	OperatorSpec operator.Spec `json:"operator_spec"`
	Next         NextTask      `json:"next,omitempty"`
}

func (t *Task) GetStatus() string {
	return t.Status
}

func (t *Task) SetStatus(status string) {
	t.Status = status
}

func (t *Task) GetMessage() string {
	return t.Message
}

func (t *Task) SetMessage(msg string) {
	t.Message = msg
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
	case SucceedStatus, FailedStatus, CanceledStatus:
		return true
	default:
		return false
	}
}
