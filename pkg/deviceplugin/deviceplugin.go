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
	"path/filepath"
	"time"

	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	dm "k8s.io/kubernetes/pkg/kubelet/cm/devicemanager"

	"github.com/fsnotify/fsnotify"
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
	filePath := filepath.Join(sInfo.HostVolumeMount, "test-file")
	klog.Infof("filePath: %v", filePath)
	// _, err := os.Stat(filePath)
	// if err != nil {
	// 	return fmt.Errorf("file does not exist: %v ", err)
	// }
	// klog.Infof("File exists: %v", filePath)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("NewWatcher failed: %v ", err)
	}
	defer watcher.Close()

	updateCh := make(chan bool)
	defer close(updateCh)
	go func() {

		klog.Infof("Starting go routine")
		for {
			klog.Infof("Waiting for a file event")

			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				klog.Infof("%s %s\n", event.Name, event.Op)
				switch {
				case event.Op&fsnotify.Write == fsnotify.Write:
					klog.Infof("Write:  %s: %s", event.Op, event.Name)
					updateCh <- true
				case event.Op&fsnotify.Create == fsnotify.Create:
					klog.Infof("Create: %s: %s", event.Op, event.Name)
					updateCh <- true
					// case event.Op&fsnotify.Remove == fsnotify.Remove:
					// 	klog.Infof("Remove: %s: %s", event.Op, event.Name)
					// 	removeCh <- true
					// case event.Op&fsnotify.Rename == fsnotify.Rename:
					// 	klog.Infof("Rename: %s: %s", event.Op, event.Name)
					// 	renameCh <- true
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				klog.Infof("error: %w", err)
			}
		}
	}()

	err = watcher.Add(filePath)
	if err != nil {
		klog.Infof("filePath: %v", filePath)
		return fmt.Errorf("add failed: %w", err)
	}

	klog.Infof("Waiting for event on updateCh")
	<-updateCh
	klog.Infof("Got event on updateCh")
	klog.Infof("file was Updated")
	if err := dp1.Register(pluginapi.KubeletSocket, sInfo.ResourceName, pluginapi.DevicePluginPath); err != nil {
		return fmt.Errorf("unable to register the DevicePlugin; error: %w", err)
	}
	// klog.Infof("Creating file after registration has succeeded %s", filepath.Join(sInfo.HostVolumeMount, "test-file"))
	// file, err := os.Create(filepath.Join(sInfo.HostVolumeMount, "test-file"))
	// if err != nil {
	// 	return fmt.Errorf("failed to create a file: %v", err)
	// }
	// defer file.Close()

	// _, err = file.WriteString("registered")
	// if err != nil {
	// 	return fmt.Errorf("failed to write to file: %v", err)
	// }
	select {}
}

func waitUntilFind(filename string) error {
	for {
		time.Sleep(1 * time.Second)
		_, err := os.Stat(filename)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				return err
			}
		}
		break
	}
	return nil
}
