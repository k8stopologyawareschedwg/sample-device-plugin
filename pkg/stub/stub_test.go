package stub_test

import (
	"github.com/k8stopologyawareschedwg/sample-device-plugin/pkg/deviceconfig"
	"github.com/k8stopologyawareschedwg/sample-device-plugin/pkg/stub"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stub tests", func() {
	When(" initialized new stub object", func() {
		devices := &deviceconfig.NodesDevices{
			Devices: map[string][]deviceconfig.SampleDevice{
				"node1": {
					{
						ID:       "2",
						Healthy:  false,
						NUMANode: 0,
					},
					{
						ID:       "3",
						Healthy:  true,
						NUMANode: 1,
					},
				},
				"node2": {
					{
						ID:       "4",
						Healthy:  true,
						NUMANode: 1,
					},
					{
						ID:       "5",
						Healthy:  false,
						NUMANode: 1,
					},
				},
			}}
		sInfo, err := stub.New("devA", devices, "dev/null", "node2")
		Expect(err).ToNot(HaveOccurred())
		It("should build the correct api plugin devices config", func() {
			Expect(sInfo.APIDevsConfig[0].ID).To(Equal("4"))
			Expect(sInfo.APIDevsConfig[1].Health).To(Equal("Unhealthy"))
			Expect(sInfo.APIDevsConfig[1].Topology.Nodes[0].ID).To(Equal(int64(1)))
		})
	})
})
