#!/usr/bin/env bash

DIRNAME="$(dirname "$(readlink -f "$0")")"

kubectl apply -f "$DIRNAME"/../config
for ds in "$DIRNAME"/../manifests/deviceplugin*-ds.yaml; do
  kubectl apply -f "$ds"
done
