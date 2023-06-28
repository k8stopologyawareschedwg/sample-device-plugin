package fakedevice

import (
	"fmt"
	"strings"

	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// MakeEnv creates the following environment variables
// - _SAMPLE_DEVICE_RESOURCE=<resource-name>
// - _SAMPLE_DEVICE_ASSIGNED=<device-id>
// - _SAMPLE_DEVICE_NUMA_LOCALITY=<numanode-id>
// - <resource-prefix>_<device-id>_<device-file>=<numanode-id>
func MakeEnv(resourceName string, dev pluginapi.Device) map[string]string {
	env := map[string]string{
		"_SAMPLE_DEVICE_RESOURCE": resourceName,
		"_SAMPLE_DEVICE_ASSIGNED": dev.ID,
	}

	key := fmt.Sprintf("%s_%s", resourceName, dev.ID)
	key = strings.Map(func(r rune) rune {
		if r == '.' || r == '/' {
			return -1
		}
		return r
	}, key)
	key = strings.ToUpper(key)

	val := "-1"
	if dev.Topology != nil && len(dev.Topology.Nodes) != 0 {
		val = fmt.Sprintf("%d", dev.Topology.Nodes[0].ID)
	}
	klog.Infof("Creating environment variables key=%q val=%q", key, val)
	env[key] = val

	env["_SAMPLE_DEVICE_NUMA_LOCALITY"] = val

	return env
}
