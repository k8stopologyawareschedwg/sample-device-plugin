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
	"os"
	"strconv"

	"k8s.io/klog/v2"

	"github.com/spf13/pflag"

	"github.com/k8stopologyawareschedwg/sample-device-plugin/pkg/deviceconfig"
	"github.com/k8stopologyawareschedwg/sample-device-plugin/pkg/deviceplugin"
	"github.com/k8stopologyawareschedwg/sample-device-plugin/pkg/stub"
)

const (
	EnvVarResourceName    = "DEVICE_RESOURCE_NAME"
	EnvVarGenerateDevices = "DEVICE_GENERATE_COUNT"
)

func main() {
	configDirPath := ""
	devResourceName := ""
	generateDevs := 0

	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.StringVarP(&configDirPath, "config-dir", "C", "", "directory which contains the device plugin configuration files")
	pflag.StringVarP(&devResourceName, "resource", "r", defaultResName(), "device plugin resource name")
	pflag.IntVarP(&generateDevs, "generate", "G", defaultGenDevs(), "autogenerate devices")
	pflag.Parse()

	var conf *deviceconfig.NodesDevices

	if generateDevs > 0 {
		conf = deviceconfig.Generate(generateDevs)
	} else {
		var err error
		conf, err = deviceconfig.Parse(configDirPath, devResourceName)
		if err != nil {
			klog.Fatalf("failed to read deviceconfig; error: %v", err)
		}
	}

	sInfo, err := stub.New(devResourceName, conf, "", "")
	if err != nil {
		klog.Fatalf("%v", err)
	}

	if err = deviceplugin.Execute(sInfo, "", false, false); err != nil {
		klog.Fatalf("%v", err)
	}
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

func defaultGenDevs() int {
	devGenAmount, ok := os.LookupEnv(EnvVarGenerateDevices)
	if !ok {
		return 0
	}
	devCount, err := strconv.Atoi(devGenAmount)
	if err != nil {
		klog.Errorf("failed to convert device count %q err %s", devGenAmount, err)
		return 0
	}
	klog.Infof("autogenerate syntetic device count %d", devCount)
	return devCount
}
