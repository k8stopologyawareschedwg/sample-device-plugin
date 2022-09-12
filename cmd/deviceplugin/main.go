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
	"flag"
	"fmt"
	"github.com/k8stopologyawareschedwg/sample-device-plugin/pkg/deviceconfig"
	"github.com/k8stopologyawareschedwg/sample-device-plugin/pkg/stub"
	"os"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	dm "k8s.io/kubernetes/pkg/kubelet/cm/devicemanager"
)

const (
	EnvVarResourceName = "DEVICE_RESOURCE_NAME"
	socketDir          = "/var/lib/kubelet/device-plugins"
)

func main() {
	configDirPath := ""
	devResourceName := ""

	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.StringVarP(&configDirPath, "config-dir", "C", "", "directory which contains the device plugin configuration files")
	pflag.StringVarP(&devResourceName, "resource", "r", defaultResName(), "device plugin resource name")
	pflag.Parse()

	conf, err := deviceconfig.Parse(configDirPath, devResourceName)
	if err != nil {
		klog.Fatalf("failed to read deviceconfig; error: %v", err)
	}

	sInfo, err := stub.New(devResourceName, conf, "", "")
	if err != nil {
		klog.Fatalf("%v", err)
	}
	klog.Infof("pluginSocksDir: %s", socketDir)

	socketPath := socketDir + "/dp." + fmt.Sprintf("%d", time.Now().Unix())

	dp1 := dm.NewDevicePluginStub(sInfo.APIDevsConfig, socketPath, sInfo.ResourceName, false, false)
	if err := dp1.Start(); err != nil {
		klog.Fatalf("Unable to start the DevicePlugin, Error: %v", err)

	}
	dp1.SetAllocFunc(sInfo.GetStubAllocFunc())
	if err := dp1.Register(pluginapi.KubeletSocket, sInfo.ResourceName, pluginapi.DevicePluginPath); err != nil {
		klog.Fatalf("Unable to register the DevicePlugin, Error: %v", err)
	}
	select {}
}

func defaultResName() string {
	devResourceName, ok := os.LookupEnv(EnvVarResourceName)
	if !ok {
		klog.Infof("no resource name configured - nothing to do")
		os.Exit(0)
	}

	klog.Infof("resource name configured from environment: %q", devResourceName)
	return devResourceName
}
