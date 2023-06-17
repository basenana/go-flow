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

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"
)

var (
	localOperatorBuilder = map[string]func(operatorSpec Spec) (Operator, error){
		ShellOperator:    newLocalShellOperator,
		PythonOperator:   newLocalPythonOperator,
		MySQLOperator:    newLocalMySQLOperator,
		PostgresOperator: newLocalPostgresOperator,
	}
	localOperatorBuilderMux sync.Mutex
)

func RegisterLocalOperatorBuilder(name string, builder func(operatorSpec Spec) (Operator, error)) error {
	localOperatorBuilderMux.Lock()
	defer localOperatorBuilderMux.Unlock()
	if _, ok := localOperatorBuilder[name]; ok {
		return OperatorIsExisted
	}
	localOperatorBuilder[name] = builder
	return nil
}

type LocalExecutor struct {
	flow *Flow
}

func (l *LocalExecutor) Setup(ctx context.Context) error {
	if err := initFlowWorkDir(LocalWorkdirBase, l.flow.ID); err != nil {
		return err
	}
	return nil
}

func (l *LocalExecutor) Teardown(ctx context.Context) {
	if err := cleanUpFlowWorkDir(LocalWorkdirBase, l.flow.ID); err != nil {
		return
	}
}

func (l *LocalExecutor) DoOperation(ctx context.Context, operatorSpec Spec) error {
	localOperatorBuilderMux.Lock()
	builder, ok := localOperatorBuilder[operatorSpec.Type]
	localOperatorBuilderMux.Unlock()
	if !ok {
		return OperatorNotFound
	}
	operator, err := builder(operatorSpec)
	if err != nil {
		return err
	}

	param := Parameter{
		FlowID:  l.flow.ID,
		Workdir: flowWorkdir(LocalWorkdirBase, l.flow.ID),
	}
	return operator.Do(ctx, param)
}

func NewLocalExecutor(flow *Flow) Executor {
	return &LocalExecutor{flow: flow}
}

type localShellOperator struct {
	command string
	spec    Spec
}

func (l *localShellOperator) Do(ctx context.Context, param Parameter) error {
	command := l.command
	if command == "" {
		command = "sh"
	}

	shellFile := fmt.Sprintf("script_%d.sh", time.Now().Unix())
	shellFilePath := path.Join(param.Workdir, shellFile)

	shF, err := os.Create(shellFilePath)
	if err != nil {
		return err
	}

	if l.spec.Script != nil && l.spec.Script.Content != "" {
		switch command {
		case "sh":
			if _, err = shF.WriteString("#!/bin/sh\nset -xe\n"); err != nil {
				return err
			}
		}
		if _, err = shF.WriteString(l.spec.Script.Content); err != nil {
			return err
		}
	}

	err = os.Chmod(shF.Name(), 0755)
	if err != nil {
		return err
	}

	var (
		stdout bytes.Buffer
		cmd    *exec.Cmd
	)
	if l.spec.Script != nil && len(l.spec.Script.Command) > 0 {
		args := l.spec.Script.Command[1:]
		cmd = exec.Command(l.spec.Script.Command[0], args...)
	} else {
		cmd = exec.Command(command, shellFilePath)
	}

	env := os.Environ()
	for k, v := range l.spec.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env
	cmd.Dir = param.Workdir
	cmd.Stdout = &stdout

	err = cmd.Run()
	return err
}

func newLocalShellOperator(operatorSpec Spec) (Operator, error) {
	return &localShellOperator{spec: operatorSpec}, nil
}

type localPythonOperator struct {
	spec Spec
}

func (l *localPythonOperator) Do(ctx context.Context, param Parameter) error {
	pythonBin := "python"
	if LocalPythonVersion == "3" {
		pythonBin = "python3"
	}
	op := localShellOperator{spec: l.spec, command: pythonBin}
	return op.Do(ctx, param)
}

func newLocalPythonOperator(operatorSpec Spec) (Operator, error) {
	return &localPythonOperator{spec: operatorSpec}, nil
}

type localMySQLOperator struct{}

func (l *localMySQLOperator) Do(ctx context.Context, param Parameter) error {
	//TODO implement me
	panic("implement me")
}

func newLocalMySQLOperator(operatorSpec Spec) (Operator, error) {
	return &localMySQLOperator{}, nil
}

type localPostgresOperator struct{}

func (l *localPostgresOperator) Do(ctx context.Context, param Parameter) error {
	//TODO implement me
	panic("implement me")
}

func newLocalPostgresOperator(operatorSpec Spec) (Operator, error) {
	return &localPostgresOperator{}, nil
}
