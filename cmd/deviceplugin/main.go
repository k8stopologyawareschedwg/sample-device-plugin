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
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/pflag"

	"github.com/swatisehgal/sample-device-plugin/pkg/server"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	EnvVarResourceName = "DEVICE_RESOURCE_NAME"
)

type deviceConfig struct {
	ID       string `json:"id"`
	Healthy  bool   `json:"healthy"`
	NUMANode int    `json:"numanode"`
}

func (dc deviceConfig) ToHealthy() string {
	if dc.Healthy {
		return pluginapi.Healthy
	}
	return pluginapi.Unhealthy
}

type pluginConfig struct {
	Devices map[string][]deviceConfig `json:"devices"`
}

const (
	socketDir = "/var/lib/kubelet/device-plugins"
)

// stubAllocFunc creates and returns allocation response for the input allocate request
func stubAllocFunc(r *pluginapi.AllocateRequest, devs map[string]pluginapi.Device) (*pluginapi.AllocateResponse, error) {
	var responses pluginapi.AllocateResponse
	i := 0
	for _, req := range r.ContainerRequests {
		response := &pluginapi.ContainerAllocateResponse{}
		var env map[string]string
		var fpath, devName string
		env = make(map[string]string)
		for _, requestID := range req.DevicesIDs {
			dev, ok := devs[requestID]
			if !ok {
				return nil, fmt.Errorf("invalid allocation request with non-existing device %s", requestID)
			}

			if dev.Health != pluginapi.Healthy {
				return nil, fmt.Errorf("invalid allocation request with unhealthy device: %s", requestID)
			}

			// create fake device file
			devName = fmt.Sprintf("tty1%d", i)
			fpath = fmt.Sprintf("/dev/%s", devName)
			i++
			key := fmt.Sprintf("NUMANODE_%s_%s", dev.ID, devName)
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

func readConfig(path string) (*pluginConfig, error) {
	src, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dec := json.NewDecoder(src)
	var conf pluginConfig
	err = dec.Decode(&conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func configFilePath(configDirPath, resourceName string) string {
	configFileName := fmt.Sprintf("%s.json", strings.Map(func(r rune) rune {
		if r == '.' || r == '/' {
			return '_'
		}
		return r
	}, resourceName))
	return filepath.Join(configDirPath, configFileName)
}

func main() {
	configDirPath := ""
	resourceName := ""

	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.StringVarP(&configDirPath, "config-dir", "C", "", "directory which contains the device plugin configuration files")
	pflag.StringVarP(&resourceName, "resource", "r", "", "device plugin resource name")
	pflag.Parse()

	if configDirPath == "" {
		klog.Infof("No config provided - nothing to do")
		os.Exit(0)
	}

	if resourceName == "" {
		resourceName = os.Getenv(EnvVarResourceName)
		klog.Infof("Resource name configured from environ: %q", resourceName)
	}
	if resourceName == "" {
		klog.Infof("No resource name configured - nothing to do")
		os.Exit(0)
	}

	hostname, err := os.Hostname()
	if err != nil {
		klog.Fatalf("Unable to get the hostname, Error: %v", err)
	}

	fullPath := configFilePath(configDirPath, resourceName)
	klog.Infof("configuration file path is %q", fullPath)
	conf, err := readConfig(fullPath)
	if err != nil {
		klog.Fatalf("Unable to read the config, Error: %v", err)
	}

	devsConf := conf.Devices[hostname]
	if len(devsConf) == 0 {
		devsConf = conf.Devices["*"]
	}
	if len(devsConf) == 0 {
		klog.Infof("No devices configured for %q - nothing to do", hostname)
		os.Exit(0)
	}

	klog.V(4).Infof("Devices config: %s", spew.Sdump(devsConf))

	var devs []*pluginapi.Device
	for _, devConf := range devsConf {
		devs = append(devs, &pluginapi.Device{
			ID:     devConf.ID,
			Health: devConf.ToHealthy(),
			Topology: &pluginapi.TopologyInfo{
				Nodes: []*pluginapi.NUMANode{
					{ID: int64(devConf.NUMANode)},
				},
			},
		})
	}

	if len(devs) == 0 {
		klog.Infof("No devices enabled for resource %q - nothing to do", resourceName)
		os.Exit(0)
	}

	klog.V(3).Infof("Devices: %s", spew.Sdump(devs))
	klog.Infof("pluginSocksDir: %s", socketDir)

	socketPath := socketDir + "/dp." + fmt.Sprintf("%d", time.Now().Unix())

	dp1 := server.NewDevicePlugin(devs, socketPath, resourceName, false)
	if err := dp1.Start(); err != nil {
		klog.Fatalf("Unable to start the DevicePlugin, Error: %v", err)

	}
	dp1.SetAllocFunc(stubAllocFunc)
	if err := dp1.Register(pluginapi.KubeletSocket, resourceName, pluginapi.DevicePluginPath); err != nil {
		klog.Fatalf("Unable to register the DevicePlugin, Error: %v", err)
	}
	select {}
}
