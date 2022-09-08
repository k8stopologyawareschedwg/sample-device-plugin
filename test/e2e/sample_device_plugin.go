package sample_device_plugin

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	DefaultDevicePath = "/dev/null"
)

type deviceRequest struct {
	Name   string
	Amount int64
}

var _ = Describe("sample device plugin", func() {
	Context("with pod requesting devices ", func() {
		var testpod *corev1.Pod

		AfterEach(func() {
			err := Client.Delete(context.TODO(), testpod)
			Expect(err).ToNot(HaveOccurred())

			err = WaitForPodDeletion(testpod, 60*time.Second)
			Expect(err).ToNot(HaveOccurred())
		})

		table.DescribeTable("should run with", func(devReqs []deviceRequest) {
			var err error
			testpod = GetTestPod()
			testpod.Namespace = TestingNamespace.Name
			if len(devReqs) > 0 {
				for idx := 0; idx < len(testpod.Spec.Containers); idx++ {
					testpod.Spec.Containers[idx].Resources.Requests = make(map[corev1.ResourceName]resource.Quantity)
					testpod.Spec.Containers[idx].Resources.Limits = make(map[corev1.ResourceName]resource.Quantity)

					for _, devReq := range devReqs {
						testpod.Spec.Containers[idx].Resources.Requests[corev1.ResourceName(devReq.Name)] = *resource.NewQuantity(devReq.Amount, resource.DecimalSI)
						testpod.Spec.Containers[idx].Resources.Limits[corev1.ResourceName(devReq.Name)] = *resource.NewQuantity(devReq.Amount, resource.DecimalSI)
					}
				}
			}

			err = Client.Create(context.TODO(), testpod)
			Expect(err).ToNot(HaveOccurred())

			err = WaitForPodCondition(testpod, corev1.PodReady, corev1.ConditionTrue, 10*time.Minute)
			Expect(err).ToNot(HaveOccurred())
			devicesMatchCount := 0
			for _, devReq := range devReqs {

				envVars, err := ExecCommandOnPod(testpod, []string{"env"})
				Expect(err).ToNot(HaveOccurred())
				for _, envVar := range strings.Split(string(envVars), "\n") {
					env := strings.TrimSpace(envVar)
					devName := strings.Map(func(r rune) rune {
						if r == '.' || r == '/' {
							return -1
						}
						return r
					}, devReq.Name)
					if len(env) == 0 || !strings.HasPrefix(env, devName) {
						continue
					}
					Expect(env).To(HavePrefix(strings.ToUpper(devName)))
				}

				data, err := ExecCommandOnPod(testpod, []string{"/bin/sh", "-c", fmt.Sprintf("/bin/stat -c %%F %s", DefaultDevicePath)})
				Expect(err).ToNot(HaveOccurred())
				for _, devDesc := range strings.Split(string(data), "\n") {
					line := strings.TrimSpace(devDesc)
					if len(line) == 0 {
						continue
					}
					Expect(devDesc).To(ContainSubstring("character special file"))
					devicesMatchCount++
				}
			}
			Expect(devicesMatchCount).To(Equal(len(devReqs)))
		},
			table.Entry("a single device A", []deviceRequest{
				{
					Name:   "example.com/deviceA",
					Amount: 1,
				},
			}),
			table.Entry("a single device B", []deviceRequest{
				{
					Name:   "example.com/deviceB",
					Amount: 1,
				},
			}),
		)
	})
})
