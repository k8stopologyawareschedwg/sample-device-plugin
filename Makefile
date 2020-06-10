COMMONENVVAR=GOOS=linux GOARCH=amd64
BUILDENVVAR=CGO_ENABLED=0

.PHONY: all
all: build

.PHONY: build
build: gofmt
	 $(COMMONENVVAR) $(BUILDENVVAR) go build -ldflags '-w' -o ./bin/devicepluginA ./cmd/devicepluginA
	 $(COMMONENVVAR) $(BUILDENVVAR) go build -ldflags '-w' -o ./bin/devicepluginB ./cmd/devicepluginB

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
	docker build -f images/Dockerfile-dpA -t quay.io/swsehgal/device-plugin-a:latest .
	docker build -f images/Dockerfile-dpB -t quay.io/swsehgal/device-plugin-b:latest .

.PHONY: push
push: image
	@echo "pushing image"
	docker push quay.io/swsehgal/device-plugin-a:latest
	docker push quay.io/swsehgal/device-plugin-b:latest

.PHONY: push-A
push-A: image
	@echo "pushing image device plugin A"
	docker push quay.io/swsehgal/device-plugin-a:latest

.PHONY: push-B
push-B: build
	@echo "pushing image device plugin B"
	docker push quay.io/swsehgal/device-plugin-b:latest


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
	@echo "Deploying device plugin B"
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

clean:
	rm -f ./bin/devicepluginA
	rm -f ./bin/devicepluginB
	kubectl delete -f manifests/devicepluginA-ds.yaml
	kubectl delete -f manifests/devicepluginB-ds.yaml
