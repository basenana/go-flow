package flow

import (
	"context"
	"github.com/google/uuid"
	"os"
	"testing"
)

func init() {
	LocalWorkdirBase = os.TempDir()
	LocalPythonVersion = "3"
}

func TestLocalExecutor_ShellOperator(t *testing.T) {
	tests := []struct {
		name string
		flow *Flow
		spec Spec
	}{
		{
			name: "test-sh-content",
			flow: &Flow{ID: uuid.New().String()},
			spec: Spec{
				Type: ShellOperator,
				Script: &Script{
					Content: "echo 'Hello World!'",
				},
				Parameter: map[string]string{},
				Env:       map[string]string{},
			},
		},
		{
			name: "test-sh-command",
			flow: &Flow{ID: uuid.New().String()},
			spec: Spec{
				Type: ShellOperator,
				Script: &Script{
					Command: []string{"sh", "-c", `echo "Hello World!"`},
				},
				Parameter: map[string]string{},
				Env:       map[string]string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLocalExecutor(tt.flow)
			ctx := context.TODO()
			if err := l.Setup(ctx); err != nil {
				t.Errorf("Setup() error = %v", err)
			}
			if err := l.DoOperation(ctx, tt.spec); err != nil {
				t.Errorf("DoOperation() error = %v", err)
			}
			l.Teardown(ctx)
		})
	}
}

func TestLocalExecutor_PythonOperator(t *testing.T) {
	tests := []struct {
		name string
		flow *Flow
		spec Spec
	}{
		{
			name: "test-python-content",
			flow: &Flow{ID: uuid.New().String()},
			spec: Spec{
				Type: PythonOperator,
				Script: &Script{
					Content: `print("Hello World!")`,
				},
				Parameter: map[string]string{},
				Env:       map[string]string{},
			},
		},
		{
			name: "test-sh-command",
			flow: &Flow{ID: uuid.New().String()},
			spec: Spec{
				Type: PythonOperator,
				Script: &Script{
					Command: []string{"python3", "-c", `print("Hello World!")`},
				},
				Parameter: map[string]string{},
				Env:       map[string]string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLocalExecutor(tt.flow)
			ctx := context.TODO()
			if err := l.Setup(ctx); err != nil {
				t.Errorf("Setup() error = %v", err)
			}
			if err := l.DoOperation(ctx, tt.spec); err != nil {
				t.Errorf("DoOperation() error = %v", err)
			}
			l.Teardown(ctx)
		})
	}
}
