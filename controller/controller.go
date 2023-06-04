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

package controller

import (
	"context"
	"fmt"
	"github.com/basenana/go-flow/executor"
	"github.com/basenana/go-flow/storage"
	"github.com/basenana/go-flow/utils"
)

type Controller struct {
	flows   map[string]executor.Runner
	storage storage.Interface
	logger  utils.Logger
}

func (c *Controller) TriggerFlow(ctx context.Context, flowId string) error {
	f, err := c.storage.GetFlow(ctx, flowId)
	if err != nil {
		return err
	}

	r := executor.NewRunner(f, c.storage)
	c.flows[f.ID] = r

	c.logger.Infof("trigger flow %s", flowId)
	return r.Start(ctx)
}

func (c *Controller) PauseFlow(flowId string) error {
	r, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	c.logger.Infof("pause flow %s", flowId)
	return r.Pause()
}

func (c *Controller) CancelFlow(flowId string) error {
	r, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	return r.Cancel()
}

func (c *Controller) ResumeFlow(flowId string) error {
	r, ok := c.flows[flowId]
	if !ok {
		return fmt.Errorf("flow %s not found", flowId)
	}
	return r.Resume()
}

func NewFlowController(opt Option) (*Controller, error) {
	return &Controller{
		flows:   make(map[string]executor.Runner),
		storage: opt.Storage,
		logger:  utils.NewLogger("go-flow"),
	}, nil
}
