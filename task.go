/*
   Copyright 2024 Go-Flow Authors

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

package go_flow

type Task interface {
	GetName() string
	GetStatue() string
	SetStatue(string)
	GetMessage() string
	SetMessage(string)
}

type BasicTask struct {
	Name    string
	Status  string
	Message string
}

func (t *BasicTask) GetStatus() string {
	return t.Status
}

func (t *BasicTask) SetStatus(status string) {
	t.Status = status
}

func (t *BasicTask) GetMessage() string {
	return t.Message
}

func (t *BasicTask) SetMessage(msg string) {
	t.Message = msg
}

type TaskDirector interface {
	Next() NextTask
}

type NextTask struct {
	OnSucceed string
	OnFailed  string
}
