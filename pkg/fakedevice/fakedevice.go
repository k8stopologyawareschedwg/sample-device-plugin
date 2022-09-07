package fakedevice

import (
	"fmt"
	"strings"

	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// MakeEnv creates the environment variable in the format: <resource-prefix>_<device-id>_<device-file>=<numanode-id>
func MakeEnv(resourceName string, dev pluginapi.Device) map[string]string {
	env := make(map[string]string)

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

	return env
}
