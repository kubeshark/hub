SHELL=/bin/bash

.PHONY: help
.DEFAULT_GOAL := build
.ONESHELL:

help: ## Print this help message.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the program.
	go build -ldflags="-extldflags=-static -s -w" -o hub .

test: ## Run the tests.
	@go test ./... -coverpkg=./... -race -coverprofile=coverage.out -covermode=atomic

lint: ## Lint the source code.
	golangci-lint run

run: ## Run the program. Requires Worker being available on port 8897
	./hub -port 8898 -debug

docker: ## Build the Docker image.
	docker build . -t ${DOCKER_REPO}:${DOCKER_TAG} --build-arg TARGETARCH=amd64

docker-push: ## Push the Docker image into Docker Hub.
	docker push ${DOCKER_REPO}:${DOCKER_TAG}

docker-latest: ## Build and push the Docker image with 'latest' tag
	export DOCKER_REPO='kubeshark/hub' && \
	export DOCKER_TAG='latest' && \
	${MAKE} docker && \
	${MAKE} docker-push

docker-canary: ## Build and push the Docker image with 'canary' tag
	export DOCKER_REPO='kubeshark/hub' && \
	export DOCKER_TAG='canary' && \
	${MAKE} docker && \
	${MAKE} docker-push

docker-dev: ## Build and push the Docker image with 'dev' tag
	export DOCKER_REPO='kubeshark/hub' && \
	export DOCKER_TAG='dev' && \
	${MAKE} docker && \
	${MAKE} docker-push
