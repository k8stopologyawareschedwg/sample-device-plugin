# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM busybox
ADD bin/deviceplugin /bin/deviceplugin
RUN mkdir -p /etc/devices
ADD config/devices.json /etc/devices/config.json
ADD config/device_A.json /etc/devices/example_com_deviceA.json
ADD config/device_B.json /etc/devices/example_com_deviceB.json
ENTRYPOINT ["/bin/deviceplugin", "--alsologtostderr", "-C", "/etc/devices"]
