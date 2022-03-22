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
	"github.com/basenana/go-flow/flow"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("test_runner", func() {
	var (
		builder = ChaosFlowBuilder{
			BatchNum: 2,
			Parallel: 2,
			Policy: flow.ControlPolicy{
				FailedPolicy: flow.PolicyFastFailed,
			},
			Chaos: ChaosPolicy{},
		}
		flowObj flow.Flow
	)
	Context("create-simple-flow", func() {
		flowObj = builder.Build()
		Expect(flowStorage.SaveFlow(flowObj)).Should(BeNil())
	})

	Context("trigger-simper-flow", func() {
		Expect(controller.TriggerFlow(context.TODO(), flowObj.ID())).Should(BeNil())
		Eventually(func() []byte {
			return status2Bytes(flowObj.GetStatus())
		}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.RunningStatus)))
	})

	Context("wait-flow-finish", func() {
		Eventually(func() []byte {
			return status2Bytes(flowObj.GetStatus())
		}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.SucceedStatus)))
	})
})

var _ = Describe("test_runner_failed", func() {

	Describe("test-fast-failed", func() {
		var (
			builder ChaosFlowBuilder
			flowObj flow.Flow
		)
		BeforeEach(func() {
			flowObj = builder.Build()
			Expect(flowStorage.SaveFlow(flowObj)).Should(BeNil())
		})

		Context("test-fast-failed-1", func() {
			builder = ChaosFlowBuilder{
				BatchNum: 3,
				Parallel: 2,
				Policy: flow.ControlPolicy{
					FailedPolicy: flow.PolicyFastFailed,
				},
				Chaos: ChaosPolicy{
					FailTaskIdx: 1,
					TaskErr:     fmt.Errorf("mocked error"),
				},
			}

			It("failed at first", func() {
				Expect(controller.TriggerFlow(context.TODO(), flowObj.ID())).Should(BeNil())
				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.FailedStatus)))
			})
		})

		Context("test-fast-failed-2", func() {
			builder = ChaosFlowBuilder{
				BatchNum: 3,
				Parallel: 2,
				Policy: flow.ControlPolicy{
					FailedPolicy: flow.PolicyFastFailed,
				},
				Chaos: ChaosPolicy{
					FailTaskIdx: 4,
					TaskErr:     fmt.Errorf("mocked error"),
				},
			}
			It("failed at mid", func() {
				Expect(controller.TriggerFlow(context.TODO(), flowObj.ID())).Should(BeNil())
				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.FailedStatus)))
			})
		})

		Context("test-fast-failed-3", func() {
			builder = ChaosFlowBuilder{
				BatchNum: 3,
				Parallel: 2,
				Policy: flow.ControlPolicy{
					FailedPolicy: flow.PolicyFastFailed,
				},
				Chaos: ChaosPolicy{
					FailTaskIdx: 6,
					TaskErr:     fmt.Errorf("mocked error"),
				},
			}
			It("failed at last", func() {
				Expect(controller.TriggerFlow(context.TODO(), flowObj.ID())).Should(BeNil())
				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.FailedStatus)))
			})
		})
	})

	Describe("test-flow-setup-failed", func() {
		var (
			builder = ChaosFlowBuilder{
				BatchNum: 3,
				Parallel: 2,
				Policy: flow.ControlPolicy{
					FailedPolicy: flow.PolicyFastFailed,
				},
				Chaos: ChaosPolicy{
					FlowSetupErr: fmt.Errorf("mocked flow setup error"),
				},
			}
			flowObj flow.Flow
		)
		BeforeEach(func() {
			flowObj = builder.Build()
			Expect(flowStorage.SaveFlow(flowObj)).Should(BeNil())
		})

		Context("test-setup-failed", func() {
			It("test-setup-failed", func() {
				Expect(controller.TriggerFlow(context.TODO(), flowObj.ID())).ShouldNot(BeNil())
				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.FailedStatus)))
			})
		})
	})

	Describe("test-task-setup-failed", func() {
		var (
			builder ChaosFlowBuilder
			flowObj flow.Flow
		)
		BeforeEach(func() {
			flowObj = builder.Build()
			Expect(flowStorage.SaveFlow(flowObj)).Should(BeNil())
		})

		Context("test-first-task-setup-failed", func() {
			builder = ChaosFlowBuilder{
				BatchNum: 3,
				Parallel: 2,
				Policy: flow.ControlPolicy{
					FailedPolicy: flow.PolicyFastFailed,
				},
				Chaos: ChaosPolicy{
					FailTaskIdx:  1,
					TaskSetupErr: fmt.Errorf("mocked setup error"),
				},
			}
			It("fist task failed", func() {
				Expect(controller.TriggerFlow(context.TODO(), flowObj.ID())).Should(BeNil())
				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.FailedStatus)))
			})
		})

		Context("test-mid-task-setup-failed", func() {
			builder = ChaosFlowBuilder{
				BatchNum: 3,
				Parallel: 2,
				Policy: flow.ControlPolicy{
					FailedPolicy: flow.PolicyFastFailed,
				},
				Chaos: ChaosPolicy{
					FailTaskIdx:  4,
					TaskSetupErr: fmt.Errorf("mocked setup error"),
				},
			}
			It("mid task failed", func() {
				Expect(controller.TriggerFlow(context.TODO(), flowObj.ID())).Should(BeNil())
				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.FailedStatus)))
			})
		})

		Context("test-last-task-setup-failed", func() {
			builder = ChaosFlowBuilder{
				BatchNum: 3,
				Parallel: 2,
				Policy: flow.ControlPolicy{
					FailedPolicy: flow.PolicyFastFailed,
				},
				Chaos: ChaosPolicy{
					FailTaskIdx:  6,
					TaskSetupErr: fmt.Errorf("mocked setup error"),
				},
			}
			It("last task failed", func() {
				Expect(controller.TriggerFlow(context.TODO(), flowObj.ID())).Should(BeNil())
				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.FailedStatus)))
			})
		})
	})

})

var _ = Describe("test_runner_control", func() {
	var (
		builder = ChaosFlowBuilder{
			BatchNum: 1,
			Parallel: 4,
			Policy: flow.ControlPolicy{
				FailedPolicy: flow.PolicyFastFailed,
			},
			Chaos: ChaosPolicy{},
		}
		flowObj flow.Flow
	)
	BeforeEach(func() {
		flowObj = builder.Build()
		Expect(flowStorage.SaveFlow(flowObj)).Should(BeNil())
	})

	Describe("test-simple-control", func() {
		It("pause and resume simple flow", func() {
			Context("trigger-flow", func() {
				Expect(controller.TriggerFlow(context.TODO(), flowObj.ID())).Should(BeNil())

				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.RunningStatus)))
			})

			Context("pause-flow", func() {
				Expect(controller.PauseFlow(flowObj.ID())).Should(BeNil())

				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.PausedStatus)))
			})

			Context("resume-and-wait-flow-running", func() {
				Expect(controller.ResumeFlow(flowObj.ID())).Should(BeNil())

				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.RunningStatus)))
			})
			Context("wait-flow-finish", func() {
				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.SucceedStatus)))
			})
		})

		It("cancel simple flow", func() {
			Context("trigger-flow", func() {
				Expect(controller.TriggerFlow(context.TODO(), flowObj.ID())).Should(BeNil())

				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.RunningStatus)))
			})

			Context("pause-flow", func() {
				Expect(controller.PauseFlow(flowObj.ID())).Should(BeNil())

				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.PausedStatus)))

			})

			Context("wait-flow-canceled", func() {
				Expect(controller.CancelFlow(flowObj.ID())).Should(BeNil())

				Eventually(func() []byte {
					return status2Bytes(flowObj.GetStatus())
				}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.CanceledStatus)))
			})
		})
	})
})

type ChaosFlowBuilder struct {
	BatchNum int
	Parallel int
	Policy   flow.ControlPolicy
	Chaos    ChaosPolicy
}

func (f ChaosFlowBuilder) Build() flow.Flow {
	return &ChaosFlow{
		id:        flow.FID(fmt.Sprintf("chaos-flow-%s", uuid.New().String())),
		batchLeft: f.BatchNum,
		parallel:  f.Parallel,
		status:    flow.CreatingStatus,
		policy:    f.Policy,
		chaos:     f.Chaos,
	}
}
