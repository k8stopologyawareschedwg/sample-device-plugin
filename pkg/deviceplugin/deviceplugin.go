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

package deviceplugin

import (
	"fmt"
	"os"
	"time"

	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	dm "k8s.io/kubernetes/pkg/kubelet/cm/devicemanager"

	"github.com/k8stopologyawareschedwg/sample-device-plugin/pkg/stub"
)

const (
	socketDir     = "/var/lib/kubelet/device-plugins"
	watchInterval = 1 * time.Second
)

func Execute(sInfo *stub.Info, socket string, preStartContainerFlag bool, getPreferredAllocationFlag bool) error {
	for {
		if socket == "" {
			klog.Infof("pluginSocksDir: %s", socketDir)
			socket = socketDir + "/dp." + fmt.Sprintf("%d", time.Now().Unix())
		}

		klog.InfoS("Creating the DevicePlugin", "socket", socket, "resourceName", sInfo.ResourceName)
		dp1 := dm.NewDevicePluginStub(sInfo.APIDevsConfig, socket, sInfo.ResourceName, preStartContainerFlag, getPreferredAllocationFlag)
		if err := dp1.Start(); err != nil {
			return fmt.Errorf("unable to start the DevicePlugin; error: %w", err)
		}

		dp1.SetAllocFunc(sInfo.GetStubAllocateFunc())

		klog.InfoS("Registering the DevicePlugin", "endpoint", pluginapi.KubeletSocket, "resourceName", sInfo.ResourceName, "devicePluginPath", pluginapi.DevicePluginPath)
		if err := dp1.Register(pluginapi.KubeletSocket, sInfo.ResourceName, pluginapi.DevicePluginPath); err != nil {
			return fmt.Errorf("unable to register the DevicePlugin; error: %w", err)
		}

		klog.InfoS("Entering the DevicePlugin wait loop", "watchInterval", watchInterval)
		running := true
		for running {
			_, err := os.Lstat(socket)
			if err != nil {
				// Socket file not found; restart server
				klog.ErrorS(err, "server endpoint not found, re-registering", "socketPath", socket)
				running = false
				continue
			}

			time.Sleep(watchInterval)
		}
		klog.InfoS("Exited the DevicePlugin wait loop")

		klog.InfoS("Stopping the DevicePlugin")
		if err := dp1.Stop(); err != nil {
			return fmt.Errorf("unable to stop the DevicePlugin; error: %w", err)
		}
	}
}
