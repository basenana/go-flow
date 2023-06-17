package flow

import (
	"reflect"
	"sort"
	"testing"
)

func TestDAG_nextBatchTasks(t *testing.T) {
	tests := []struct {
		name        string
		dag         *DAG
		taskBatches [][]string
	}{
		{
			name: "default-1",
			dag: mustBuildDAG([]Task{
				{Name: "task-1", Next: NextTask{}, ActiveOnFailure: false},
				{Name: "task-2", Next: NextTask{}, ActiveOnFailure: false},
				{Name: "task-3", Next: NextTask{}, ActiveOnFailure: false},
			}),
			taskBatches: [][]string{{"task-1", "task-2", "task-3"}},
		},
		{
			name: "default-2",
			dag: mustBuildDAG([]Task{
				{Name: "task-1", Next: NextTask{}, ActiveOnFailure: false},
				{Name: "task-2", Next: NextTask{}, ActiveOnFailure: false},
				{Name: "task-3", Next: NextTask{}, ActiveOnFailure: true},
			}),
			taskBatches: [][]string{{"task-1", "task-2"}},
		},
		{
			name: "sequential-1",
			dag: mustBuildDAG([]Task{
				{Name: "task-1", Next: NextTask{OnSucceed: "task-2"}, ActiveOnFailure: false},
				{Name: "task-2", Next: NextTask{OnSucceed: "task-3"}, ActiveOnFailure: false},
				{Name: "task-3", Next: NextTask{}, ActiveOnFailure: false},
			}),
			taskBatches: [][]string{{"task-1"}, {"task-2"}, {"task-3"}},
		},
		{
			name: "sequential-2",
			dag: mustBuildDAG([]Task{
				{Name: "task-1", Next: NextTask{OnSucceed: "task-3", OnFailed: "task-2"}, ActiveOnFailure: false},
				{Name: "task-2", Next: NextTask{}, ActiveOnFailure: true},
				{Name: "task-3", Next: NextTask{}, ActiveOnFailure: false},
			}),
			taskBatches: [][]string{{"task-1"}, {"task-3"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batchNum := 0
			for {
				taskBatch := tt.dag.nextBatchTasks()
				if taskBatch == nil {
					break
				}
				if !isMarchBatchTasks(taskBatch, tt.taskBatches[batchNum]) {
					t.Errorf("nextBatchTasks() = %v, want %v", taskBatch, tt.taskBatches[batchNum])
				}
				for _, t := range taskBatch {
					tt.dag.updateTaskStatus(t.taskName, SucceedStatus)
				}
				batchNum += 1
			}
		})
	}
}

func isMarchBatchTasks(crt []taskToward, expect []string) bool {
	var current []string
	for _, t := range crt {
		current = append(current, t.taskName)
	}

	sort.Strings(current)
	sort.Strings(expect)
	return reflect.DeepEqual(current, expect)
}

func mustBuildDAG(tasks []Task) *DAG {
	dag, err := buildDAG(tasks)
	if err != nil {
		panic(err)
	}
	return dag
}
