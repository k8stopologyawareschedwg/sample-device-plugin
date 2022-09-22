COMMONENVVAR = GOOS=linux GOARCH=amd64
BUILDENVVAR = CGO_ENABLED=0
RUNTIME ?= podman
REPOOWNER ?= k8stopologyawareschedwg
IMAGENAME ?= sample-device-plugin
IMAGETAG ?= latest
SAMPLE_DEVICE_PLUGIN_CONTAINER_IMAGE ?= quay.io/${REPOOWNER}/${IMAGENAME}/${IMAGETAG}
KUSTOMIZE_DEPLOY_DIR ?= default

.PHONY: all
all: build

.PHONY: build
build: gofmt
	 $(COMMONENVVAR) $(BUILDENVVAR) go build -ldflags '-w' -o ./bin/deviceplugin ./cmd/deviceplugin

.PHONY: gofmt
gofmt:
	@echo "Running gofmt"
	gofmt -s -w `find . -path ./vendor -prune -o -type f -name '*.go' -print`

.PHONY: govet
govet:
	@echo "Running go vet"
	go vet ./cmd/... ./pkg/...

outdir:
	@mkdir -p _out || :

.PHONY: image
image: build
	@echo "building image"
	$(RUNTIME) build -f images/Dockerfile -t $(SAMPLE_DEVICE_PLUGIN_CONTAINER_IMAGE) .

.PHONY: unit-tests
unit-tests:
	@echo "running unit tests"
	go test -v ./cmd/... ./pkg/...

.PHONY: push
push: image
	@echo "pushing image"
	$(RUNTIME) push $(SAMPLE_DEVICE_PLUGIN_CONTAINER_IMAGE)

.PHONY: deploy
deploy:
	@echo "Deploying device plugins"
	kubectl apply -k deployment/overlays/$(KUSTOMIZE_DEPLOY_DIR)

.PHONY: undeploy
undeploy:
	@echo "Removing device plugins"
	kubectl delete -k deployment/overlays/$(KUSTOMIZE_DEPLOY_DIR)

.PHONY: build-e2e
build-e2e: outdir
	@echo "Building E2E tests"
	go test -v -c -o _out/device-plugin-e2e.test ./test/e2e/...

.PHONY: e2e-test
e2e-test: build-e2e
	@echo "Running E2E tests"
	_out/device-plugin-e2e.test -ginkgo.v

.PHONY: test-both
test-both:
	kubectl create -f manifests/test-both-success.yaml

.PHONY: deploy-A
deploy-A:
	@echo "Deploying device plugin A"
	kubectl apply -k deployment/overlays/$(KUSTOMIZE_DEPLOY_DIR)/devicepluginA

.PHONY: test-A
test-A:
	kubectl create -f manifests/test-deviceA.yaml


.PHONY: deploy-B
deploy-B:
	@echo "Deploying device plugin B"
	kubectl apply -k deployment/overlays/$(KUSTOMIZE_DEPLOY_DIR)/devicepluginB

.PHONY: test-B
test-B:
	kubectl create -f manifests/test-deviceB.yaml

clean-binaries:
	rm -f ./bin/deviceplugin

clean: clean-binaries
	kubectl delete -f manifests/devicepluginA-ds.yaml
	kubectl delete -f manifests/devicepluginB-ds.yaml
