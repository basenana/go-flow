package main

import (
	"context"
	"github.com/zwwhdls/go-flow/controller"
	"github.com/zwwhdls/go-flow/ext"
	"github.com/zwwhdls/go-flow/flow"
	"github.com/zwwhdls/go-flow/storage"
	"time"
)

func main() {
	// 1. new storage: inmemory or etcd
	opt := controller.Option{
		Storage: storage.NewInMemoryStorage(),
	}
	ctl, err := controller.NewFlowController(opt)
	if err != nil {
		panic(err)
	}
	// 2. register flow struct
	if err := ctl.Register(&ext.GraphFlow{}); err != nil {
		panic(err)
	}

	// 3. define your own flow
	task1 := ext.NewGraphTask("task1", "appId1", "appId1")
	task2 := ext.NewGraphTask("task2", "appId2", "appId2")
	task3 := ext.NewGraphTask("task3", "appId3", "appId3")
	task4 := ext.NewGraphTask("task4", "appId4", "appId4")

	tasks := []*ext.GraphTask{task1, task2, task3, task4}
	dep := ext.NewTaskDep()
	dep.RunOrder(task1.Name(), []flow.TName{task3.Name(), task4.Name()})
	dep.RunOrder(task2.Name(), []flow.TName{task3.Name(), task4.Name()})
	dep.RunOrder(task3.Name(), []flow.TName{task4.Name()})
	//dep.RunOrder(task4.Name(), []flow.TName{task3.Name()})

	// 4. new flow in storage
	f, err := ext.NewGraphFlow(opt.Storage, tasks, dep, flow.ControlPolicy{FailedPolicy: flow.PolicyFastFailed})
	if err != nil {
		panic(err)
	}

	// 5. trigger flow
	if err := ctl.TriggerFlow(context.TODO(), f.ID()); err != nil {
		panic(err)
	}
	time.Sleep(5 * time.Minute)
}
