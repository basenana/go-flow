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

package executor

import (
	"github.com/basenana/go-flow/flow"
	"github.com/basenana/go-flow/utils"
)

type taskToward struct {
	taskName  string
	status    string
	onSucceed string
	onFailed  string
}

type DAG struct {
	tasks []taskToward
}

func (g *DAG) UpdateTaskStatus(taskName, status string) {
	for i, t := range g.tasks {
		if t.taskName != taskName {
			continue
		}
		g.tasks[i].status = status
		break
	}
	return
}

func (g *DAG) buildBatches() [][]taskToward {
	return nil
}

func (g *DAG) nextBatchTasks() []taskToward {
	batches := g.buildBatches()
	for i := range batches {
		batch := batches[i]
		allFinish := true
		for _, t := range batch {
			if !flow.IsFinishedStatus(t.status) {
				allFinish = false
				break
			}
		}
		if !allFinish {
			return batch
		}
	}
	return nil
}

type taskDep struct {
	taskSet   utils.StringSet
	taskEdges map[string][]string
	preCount  map[string]int
}

func (t *taskDep) RunOrder(firstTask string, nextTasks []string) {
	t.taskSet.Insert(firstTask)
	for _, task := range nextTasks {
		t.taskSet.Insert(task)

		edges, ok := t.taskEdges[firstTask]
		if !ok {
			edges = make([]string, 0)
		}

		exist := false
		for _, d := range edges {
			if d == task {
				exist = true
				break
			}
		}
		if exist {
			continue
		}
		edges = append(edges, task)
		t.taskEdges[firstTask] = edges
		t.preCount[task] += 1
	}
}

func buildTaskDep(tasks []taskToward) taskDep {
	return taskDep{}
}
