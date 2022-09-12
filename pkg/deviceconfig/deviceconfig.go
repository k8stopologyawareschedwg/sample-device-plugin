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

package deviceconfig

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type SampleDevice struct {
	ID       string `yaml:"id"`
	Healthy  bool   `yaml:"healthy"`
	NUMANode int    `yaml:"numanode"`
}

func (dc *SampleDevice) ToHealthy() string {
	if dc.Healthy {
		return pluginapi.Healthy
	}
	return pluginapi.Unhealthy
}

type NodesDevices struct {
	Devices map[string][]SampleDevice `yaml:"devices"`
}

func Parse(path, resName string) (*NodesDevices, error) {
	var conf *NodesDevices
	if path == "" {
		return nil, fmt.Errorf("deviceconfig path must be provided - nothing to do")
	}
	fullPath := getDevPath(path, resName)

	b, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func getDevPath(configDirPath, resourceName string) string {
	configFileName := fmt.Sprintf("%s.yaml", strings.Map(func(r rune) rune {
		if r == '.' || r == '/' {
			return '_'
		}
		return r
	}, resourceName))
	return filepath.Join(configDirPath, configFileName)
}
