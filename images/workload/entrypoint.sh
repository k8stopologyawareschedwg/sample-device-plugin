#!/bin/bash

# 1. validate resource. We can't predict, so we only check something was set, anything is fine
if [ -z "${_SAMPLE_DEVICE_RESOURCE}" ]; then
	echo "missing resource"
	exit 1
fi

# 2. validate device assignment
if [ -z "${_SAMPLE_DEVICE_ASSIGNED}" ]; then
	echo "missing device"
	exit 2
fi

DEV=$( echo ${_SAMPLE_DEVICE_ASSIGNED} | egrep '^Dev-[0-9]+$' )
if [ -z "${DEV}" ]; then
	echo "malformed device: ${_SAMPLE_DEVICE_ASSIGNED}"
	exit 4
fi

# 3. print validated device - in a k8s compatible way
echo stub devices: ${DEV}

# 4. sleep forever if we got this far
sleep inf
