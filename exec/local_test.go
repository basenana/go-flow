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

package exec

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/basenana/go-flow/cfg"
	"github.com/basenana/go-flow/flow"
)

func init() {
	cfg.LocalWorkdirBase = os.TempDir()
	cfg.LocalPythonVersion = "3"
}

func TestLocalExecutor_ShellOperator(t *testing.T) {
	tests := []struct {
		name string
		flow *flow.Flow
		spec flow.Spec
	}{
		{
			name: "test-sh-content",
			flow: &flow.Flow{ID: uuid.New().String()},
			spec: flow.Spec{
				Type: ShellOperator,
				Script: &flow.Script{
					Content: "echo 'Hello World!'",
					Env:     map[string]string{},
				},
			},
		},
		{
			name: "test-sh-command",
			flow: &flow.Flow{ID: uuid.New().String()},
			spec: flow.Spec{
				Type: ShellOperator,
				Script: &flow.Script{
					Command: []string{"sh", "-c", `echo "Hello World!"`},
					Env:     map[string]string{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLocalExecutor(tt.flow, NewLocalOperatorBuilderRegister())
			ctx := context.TODO()
			if err := l.Setup(ctx); err != nil {
				t.Errorf("Setup() error = %v", err)
			}
			if err := l.DoOperation(ctx, flow.Task{}, tt.spec); err != nil {
				t.Errorf("DoOperation() error = %v", err)
			}
			l.Teardown(ctx)
		})
	}
}

func TestLocalExecutor_PythonOperator(t *testing.T) {
	tests := []struct {
		name string
		flow *flow.Flow
		spec flow.Spec
	}{
		{
			name: "test-python-content",
			flow: &flow.Flow{ID: uuid.New().String()},
			spec: flow.Spec{
				Type: PythonOperator,
				Script: &flow.Script{
					Content: `print("Hello World!")`,
					Env:     map[string]string{},
				},
			},
		},
		{
			name: "test-sh-command",
			flow: &flow.Flow{ID: uuid.New().String()},
			spec: flow.Spec{
				Type: PythonOperator,
				Script: &flow.Script{
					Command: []string{"python3", "-c", `print("Hello World!")`},
					Env:     map[string]string{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLocalExecutor(tt.flow, NewLocalOperatorBuilderRegister())
			ctx := context.TODO()
			if err := l.Setup(ctx); err != nil {
				t.Errorf("Setup() error = %v", err)
			}
			if err := l.DoOperation(ctx, flow.Task{}, tt.spec); err != nil {
				t.Errorf("DoOperation() error = %v", err)
			}
			l.Teardown(ctx)
		})
	}
}
