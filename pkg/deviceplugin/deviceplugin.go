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
	"time"

	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	dm "k8s.io/kubernetes/pkg/kubelet/cm/devicemanager"

	"github.com/k8stopologyawareschedwg/sample-device-plugin/pkg/stub"
)

const socketDir = "/var/lib/kubelet/device-plugins"

func Execute(sInfo *stub.Info, socket string, preStartContainerFlag bool, getPreferredAllocationFlag bool) error {
	if socket == "" {
		klog.Infof("pluginSocksDir: %s", socketDir)
		socket = socketDir + "/dp." + fmt.Sprintf("%d", time.Now().Unix())
	}

	dp1 := dm.NewDevicePluginStub(sInfo.APIDevsConfig, socket, sInfo.ResourceName, preStartContainerFlag, getPreferredAllocationFlag)
	if err := dp1.Start(); err != nil {
		return fmt.Errorf("unable to start the DevicePlugin; error: %w", err)
	}

	dp1.SetAllocFunc(sInfo.GetStubAllocateFunc())
	if err := dp1.Register(pluginapi.KubeletSocket, sInfo.ResourceName, pluginapi.DevicePluginPath); err != nil {
		return fmt.Errorf("unable to register the DevicePlugin; error: %w", err)
	}
	select {}
}
