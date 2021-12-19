package controller

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/zwwhdls/go-flow/flow"
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
		err     error
	)
	Context("create-simple-flow", func() {
		flowObj, err = controller.NewFlow(builder)
		Expect(err).Should(BeNil())
	})

	Context("trigger-simper-flow", func() {
		Expect(controller.TriggerFlow(context.TODO(), flowObj.ID())).Should(BeNil())
	})

	Eventually(func() []byte {
		return status2Bytes(flowObj.GetStatus())
	}, time.Minute*3, time.Second).Should(Equal(status2Bytes(flow.SucceedStatus)))
})
