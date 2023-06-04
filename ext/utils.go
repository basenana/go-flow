/*
   Copyright 2022 Go-Flow Authors

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

package ext

import (
	"github.com/basenana/go-flow/flow"
)

type basic struct {
	id      string
	status  string
	message string
	policy  flow.ControlPolicy
}

func (b basic) ID() string {
	return b.id
}

func (b basic) GetStatus() string {
	return b.status
}

func (b *basic) SetStatus(status string) {
	b.status = status
}

func (b basic) GetMessage() string {
	return b.message
}

func (b *basic) SetMessage(msg string) {
	b.message = msg
}

func (b basic) GetControlPolicy() flow.ControlPolicy {
	return b.policy
}

type TaskNameSet map[string]struct{}

func (s TaskNameSet) Insert(t string) {
	s[t] = struct{}{}
}
func (s TaskNameSet) Has(task string) bool {
	_, ok := s[task]
	return ok
}

func (s TaskNameSet) Del(t string) {
	if _, ok := s[t]; ok {
		delete(s, t)
	}
}

func (s TaskNameSet) List() (result []string) {
	for k := range s {
		result = append(result, k)
	}
	return
}

func (s TaskNameSet) Len() int {
	return len(s)
}

func NewTaskNameSet() TaskNameSet {
	return map[string]struct{}{}
}
