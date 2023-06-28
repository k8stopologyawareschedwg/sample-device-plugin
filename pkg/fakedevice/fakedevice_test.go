package fakedevice_test

import (
	"strings"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/k8stopologyawareschedwg/sample-device-plugin/pkg/fakedevice"
)

const (
	resourceName       = "io.openshift/fakedev"
	resourcePrefix     = "IOOPENSHIFTFAKEDEV"
	sampleDevicePrefix = "_SAMPLE_DEVICE"
)

var _ = Describe("FakeDevice", func() {
	Describe("creating environment variable", func() {
		It("should have all caps name", func() {
			dev := pluginapi.Device{
				ID:       "test00",
				Health:   pluginapi.Healthy,
				Topology: &pluginapi.TopologyInfo{Nodes: []*pluginapi.NUMANode{{ID: 0}}},
			}
			env := fakedevice.MakeEnv(resourceName, dev)
			Expect(env).To(Not(BeNil()))
			for key := range env {
				keyUpper := strings.ToUpper(key)
				Expect(key).To(Equal(keyUpper))
			}
		})

		It("should not have slashes in the name", func() {
			dev := pluginapi.Device{
				ID:       "test01",
				Health:   pluginapi.Healthy,
				Topology: &pluginapi.TopologyInfo{Nodes: []*pluginapi.NUMANode{{ID: 0}}},
			}
			env := fakedevice.MakeEnv(resourceName, dev)
			Expect(env).To(Not(BeNil()))
			for key := range env {
				Expect(key).To(Not(ContainSubstring("/")))
			}
		})

		It("should not have dots in the name", func() {
			dev := pluginapi.Device{
				ID:       "test02",
				Health:   pluginapi.Healthy,
				Topology: &pluginapi.TopologyInfo{Nodes: []*pluginapi.NUMANode{{ID: 0}}},
			}
			env := fakedevice.MakeEnv(resourceName, dev)
			Expect(env).To(Not(BeNil()))
			for key := range env {
				Expect(key).To(Not(ContainSubstring(".")))
			}
		})

		It("should have Resource Name prefix", func() {
			dev := pluginapi.Device{
				ID:       "test03",
				Health:   pluginapi.Healthy,
				Topology: &pluginapi.TopologyInfo{Nodes: []*pluginapi.NUMANode{{ID: 0}}},
			}
			env := fakedevice.MakeEnv(resourceName, dev)
			Expect(env).To(Not(BeNil()))
			for key := range env {
				Expect(key).To(Or(
					HavePrefix(resourcePrefix),
					HavePrefix(sampleDevicePrefix),
				))
			}
		})
	})
})
