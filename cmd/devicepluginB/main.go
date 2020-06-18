/*
Copyright 2018 The Kubernetes Authors.

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

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/swatisehgal/sample-device-plugin/pkg/server"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	socketDir    = "/var/lib/kubelet/device-plugins"
	resourceName = "example.com/deviceB"
	masterName   = "kind-kubetest-control-plane"
)

// stubAllocFunc creates and returns allocation response for the input allocate request
func stubAllocFunc(r *pluginapi.AllocateRequest, devs map[string]pluginapi.Device) (*pluginapi.AllocateResponse, error) {
	var responses pluginapi.AllocateResponse
	i := 0
	for _, req := range r.ContainerRequests {
		response := &pluginapi.ContainerAllocateResponse{}
		var env map[string]string
		env = make(map[string]string)
		var fpath string
		for _, requestID := range req.DevicesIDs {
			dev, ok := devs[requestID]
			if !ok {
				return nil, fmt.Errorf("invalid allocation request with non-existing device %s", requestID)
			}

			if dev.Health != pluginapi.Healthy {
				return nil, fmt.Errorf("invalid allocation request with unhealthy device: %s", requestID)
			}

			fpath = fmt.Sprintf("/dev/tty2%d", i)
			i++

			key := fmt.Sprintf("%s_%s_%s", resourceName, dev.ID, fpath)
			val := fmt.Sprintf("%d", dev.Topology.Nodes[0].ID)
			klog.Infof("Creating environment variables key: %s:val %s", key, val)
			env[key] = val

			response.Devices = append(response.Devices, &pluginapi.DeviceSpec{
				ContainerPath: fpath,
				HostPath:      fpath,
				Permissions:   "rw",
			})
		}
		response.Envs = env
		responses.ContainerResponses = append(responses.ContainerResponses, response)
	}

	return &responses, nil
}

func main() {

	hostname, err := os.Hostname()
	if err != nil {
		klog.Infof("Unable to get the hostname, Error: %v", err)
	}

	var devs []*pluginapi.Device
	if hostname == masterName {
		devs = []*pluginapi.Device{
			{ID: "DevB1", Health: pluginapi.Healthy, Topology: &pluginapi.TopologyInfo{Nodes: []*pluginapi.NUMANode{{ID: 1}}}},
			{ID: "DevB2", Health: pluginapi.Healthy, Topology: &pluginapi.TopologyInfo{Nodes: []*pluginapi.NUMANode{{ID: 1}}}},
			{ID: "DevB3", Health: pluginapi.Healthy, Topology: &pluginapi.TopologyInfo{Nodes: []*pluginapi.NUMANode{{ID: 0}}}},
		}
	} else {
		devs = []*pluginapi.Device{
			{ID: "DevB1", Health: pluginapi.Healthy, Topology: &pluginapi.TopologyInfo{Nodes: []*pluginapi.NUMANode{{ID: 1}}}},
			{ID: "DevB2", Health: pluginapi.Healthy, Topology: &pluginapi.TopologyInfo{Nodes: []*pluginapi.NUMANode{{ID: 1}}}},
			{ID: "DevB3", Health: pluginapi.Healthy, Topology: &pluginapi.TopologyInfo{Nodes: []*pluginapi.NUMANode{{ID: 1}}}},
		}
	}

	klog.Infof("pluginSocksDir: %s", socketDir)

	socketPath := socketDir + "/dp." + fmt.Sprintf("%d", time.Now().Unix())

	dp1 := server.NewDevicePlugin(devs, socketPath, resourceName, false)
	if err := dp1.Start(); err != nil {
		panic(err)

	}
	dp1.SetAllocFunc(stubAllocFunc)
	if err := dp1.Register(pluginapi.KubeletSocket, resourceName, pluginapi.DevicePluginPath); err != nil {
		panic(err)
	}
	select {}
}
