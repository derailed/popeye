NAME     := popeye
PACKAGE  := github.com/derailed/$(NAME)
VERSION  := v0.22.1
GIT      := $(shell git rev-parse --short HEAD)
DATE     := $(shell date +%FT%T%Z)
IMG_NAME := derailed/popeye
IMAGE    ?= ${IMG_NAME}:${VERSION}
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le

default: help

test:      ## Run all tests
	@go clean --testcache
	@go test ./...

cover:     ## Run test coverage suite
	@go test ./... --coverprofile=cov.out
	@go tool cover --html=cov.out

build:     ## Builds the CLI
	@go build \
	-ldflags "-w -X ${PACKAGE}/cmd.version=${VERSION} -X ${PACKAGE}/cmd.commit=${GIT} -X ${PACKAGE}/cmd.date=${DATE}" \
	-a -tags netgo -o execs/${NAME} *.go

img:  ## Build Docker Image
	@docker build --rm -t ${IMAGE} .

push: img ## Push Docker Image
	@docker push ${IMAGE}

buildx: ## Build and push docker image for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- docker buildx create --name popeye-builder
	docker buildx use popeye-builder
	- docker buildx build --push --platform=$(PLATFORMS) --tag ${IMAGE} -f Dockerfile.cross .
	- docker buildx rm popeye-builder
	rm Dockerfile.cross

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[38;5;69m%-30s\033[38;5;38m %s\033[0m\n", $$1, $$2}'
