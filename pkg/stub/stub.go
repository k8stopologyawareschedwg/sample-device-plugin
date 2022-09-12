/*
Copyright 2022 The Kubernetes Authors.

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

package stub

import (
	"fmt"
	"os"

	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"github.com/k8stopologyawareschedwg/sample-device-plugin/pkg/deviceconfig"
	"github.com/k8stopologyawareschedwg/sample-device-plugin/pkg/fakedevice"
)

const DefaultDevicePath = "/dev/null"

type Info struct {
	ResourceName  string
	APIDevsConfig []*pluginapi.Device
	devicePath    string
	nodeName      string
}

// GetStubAllocateFunc creates and returns allocation response for the input allocate request
func (sInfo *Info) GetStubAllocateFunc() func(r *pluginapi.AllocateRequest, devs map[string]pluginapi.Device) (*pluginapi.AllocateResponse, error) {
	return func(r *pluginapi.AllocateRequest, devs map[string]pluginapi.Device) (*pluginapi.AllocateResponse, error) {
		var responses pluginapi.AllocateResponse
		for _, req := range r.ContainerRequests {
			response := &pluginapi.ContainerAllocateResponse{}
			env := make(map[string]string)

			for _, requestID := range req.DevicesIDs {
				dev, ok := devs[requestID]
				if !ok {
					return nil, fmt.Errorf("invalid allocation request with non-existing device %s", requestID)
				}

				if dev.Health != pluginapi.Healthy {
					return nil, fmt.Errorf("invalid allocation request with unhealthy device: %s", requestID)
				}

				for key, val := range fakedevice.MakeEnv(sInfo.ResourceName, dev) {
					env[key] = val
				}

				response.Devices = append(response.Devices, &pluginapi.DeviceSpec{
					ContainerPath: DefaultDevicePath,
					HostPath:      DefaultDevicePath,
					Permissions:   "rw",
				})
			}
			response.Envs = env
			responses.ContainerResponses = append(responses.ContainerResponses, response)
		}
		return &responses, nil
	}
}

func New(resourceName string, nodeDevicesConfig *deviceconfig.NodesDevices, devicePath, nodeName string) (*Info, error) {
	if devicePath == "" {
		klog.Infof("using default device type path: %q", DefaultDevicePath)
		devicePath = DefaultDevicePath
	}

	if nodeName == "" {
		var err error
		nodeName, err = os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("unable to get the hostname, error: %v", err)
		}
	}

	devsConf, ok := nodeDevicesConfig.Devices[nodeName]
	if !ok {
		devsConf, ok = nodeDevicesConfig.Devices["*"]
	}
	if !ok {
		return nil, fmt.Errorf("no devices configured for %q - nothing to do", nodeName)
	}
	klog.V(4).Infof("node: %q configured devices: %v", nodeName, devsConf)

	devs := MakePluginApiDevices(devsConf)
	if len(devs) == 0 {
		return nil, fmt.Errorf("no devices enabled for resource %q - nothing to do", resourceName)
	}
	klog.V(4).Infof("devices: %v", devs)

	return &Info{
		ResourceName:  resourceName,
		APIDevsConfig: devs,
		devicePath:    devicePath,
		nodeName:      nodeName,
	}, nil
}

func MakePluginApiDevices(devsConf []deviceconfig.SampleDevice) []*pluginapi.Device {
	var devs []*pluginapi.Device
	for _, devConf := range devsConf {
		var topo *pluginapi.TopologyInfo
		if devConf.NUMANode != -1 {
			topo = &pluginapi.TopologyInfo{
				Nodes: []*pluginapi.NUMANode{
					{ID: int64(devConf.NUMANode)},
				},
			}
		}
		dev := &pluginapi.Device{
			ID:       devConf.ID,
			Health:   devConf.ToHealthy(),
			Topology: topo,
		}
		devs = append(devs, dev)
	}
	return devs
}
