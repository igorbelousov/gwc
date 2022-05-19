SHELL := /bin/bash

run:
	go run app/cmd/main.go | go run app/tooling/logfmt/main.go 

build:
	go build -o "services" -ldflags "-X main.build=local"



# Building container docker 
VERSION := 0.1
SERVICE_NAME := gwc-app-amd64

all: service

service:
	docker build \
		-f zarf/docker/dockerfile.service \
		-t $(SERVICE_NAME):$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"`\
		.

# Start Kubernetes cluster

KIND_CLUSTER := gwc-claster

kind-up:
	kind create cluster \
		--image kindest/node:v1.23.0@sha256:49824ab1727c04e56a21a5d8372a402fcd32ea51ac96a2706a12af38934f81ac \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=service-system


kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

kind-load:
	
	kind load docker-image $(SERVICE_NAME):$(VERSION) --name $(KIND_CLUSTER)

kind-apply:
	kustomize build zarf/k8s/kind/service-pod/ | kubectl apply -f - 

kind-restart:
	kubectl rollout restart deployment service-pod

kind-update: all kind-load kind-restart

kind-update-apply: all kind-load kind-apply


kind-status:
	kubectl get nodes -o wide 
	kubectl get svc -o wide 
	kubectl get pods -o wide --watch --all-namespaces


kind-status-service:
	kubectl get pods -o wide --watch


kind-logs:
	kubectl logs -l app=service --all-containers=true -f --tail=100

kind-describe:
	# kubectl describe nodes
	# kubectl describe scv
	kubectl describe pod -l app=service

tidy:
	go mod tidy
	# go mod vendor

