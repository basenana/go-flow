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

import "context"

type Executor interface {
	Setup(ctx context.Context) error
	Exec(ctx context.Context, flow *Flow, task Task) (Result, error)
	Teardown(ctx context.Context) error
}

type Result struct {
	IsSucceed bool
	Retry     bool
}

type simpleExecutor struct{}

var _ Executor = &simpleExecutor{}

func (s *simpleExecutor) Exec(ctx context.Context, flow *Flow, task Task) (Result, error) {
	return Result{}, nil
}

func (s *simpleExecutor) Setup(ctx context.Context) error {
	return nil
}

func (s *simpleExecutor) Teardown(ctx context.Context) error {
	return nil
}
