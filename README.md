# Sample Device Plugin

This is a sample device plugin repository to enable support for sample devices in a kubernetes cluster using the device plugin API. The motivation for this was to emulated devices on a NUMA node basis in a kubernetes environment in case the environment does not have multi numa hardware or for testing purpose. This repo contains two device plugins to be deployed on a two node cluster (one master and one worker node). The devices emulated on various NUMA nodes on a two node cluster is explained in the diagram below:

![Setup](numa-topology.png)

## Installation

1. Update the image name and/or docker repository in the Makefile
2. To deploy the device plugin run:

```bash
make push
make deploy
```
The Makefile provides other targets:
* build: Build the device plugin go code
* gofmt: To format the code
* push: To push the docker image to a registry
* images: To build the docker image

NOTE: The Makefile also contains individual device plugin specific targets in case they need to deployed independently

## Workload requesting devices

To test the working of the device plugins, deploy test deployment that requests both devices
```python
make test-both
```
In case topology manager has been enabled with a single-numa-node policy, for a workload requesting an instance each of both devices (as in manifests/test-both-success.yaml) the devices would have to be allocated from the same NUMA node which would be only possible if the workload gets placed on the master. The nodeselector to force the pod to run on the master is to ensure that the pod is deployed successfully as the scheduler lacks topology information.
manifests/test-both-taerror.yaml showcases a scenario where Topology affinity error is caused due to failure to align resources on the same NUMA node
