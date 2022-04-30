NAME     := popeye
PACKAGE  := github.com/derailed/$(NAME)
VERSION  := v0.10.0
GIT      := $(shell git rev-parse --short HEAD)
DATE     := $(shell date +%FT%T%Z)
IMG_NAME := derailed/popeye
IMAGE    := ${IMG_NAME}:${VERSION}

default: help

test:      ## Run all tests
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

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[38;5;69m%-30s\033[38;5;38m %s\033[0m\n", $$1, $$2}'
