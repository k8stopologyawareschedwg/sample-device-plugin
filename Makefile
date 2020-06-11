COMMONENVVAR = GOOS=linux GOARCH=amd64
BUILDENVVAR = CGO_ENABLED=0
RUNTIME ?= podman
REPOOWNER ?= swsehgal
IMAGENAME ?= device-plugin
IMAGETAG ?= latest

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
	go vet

.PHONY: image
image: build
	@echo "building image"
	$(RUNTIME) build -f images/Dockerfile -t quay.io/$(REPOOWNER)/$(IMAGENAME):$(IMAGETAG) .

.PHONY: push
push: image
	@echo "pushing image"
	$(RUNTIME) push quay.io/$(REPOOWNER)/$(IMAGENAME):$(IMAGETAG)

.PHONY: deploy
deploy:
	@echo "Deploying device plugins"
	kubectl create -f manifests/devicepluginA-ds.yaml
	kubectl create -f manifests/devicepluginB-ds.yaml

.PHONY: test-both
test-both:
	kubectl create -f manifests/test-both-success.yaml


.PHONY: deploy-A
deploy-A:
	@echo "Deploying device plugin A"
	kubectl create -f manifests/devicepluginA-ds.yaml

.PHONY: test-A
test-A:
	kubectl create -f manifests/test-deviceA.yaml


.PHONY: deploy-B
deploy-B:
	@echo "Deploying device plugin B"
	kubectl create -f manifests/devicepluginB-ds.yaml

.PHONY: test-B
test-B:
	kubectl create -f manifests/test-deviceB.yaml

clean-binaries:
	rm -f ./bin/deviceplugin

clean: clean-binaries
	kubectl delete -f manifests/devicepluginA-ds.yaml
	kubectl delete -f manifests/devicepluginB-ds.yaml
