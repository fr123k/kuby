.PHONY: build

GO_VERSION := 1.13
DOCKER_CMD := docker run --rm -v $(shell pwd):/usr/src/vmv -w /usr/src/vmv -e GOPROXY=https://proxy.golang.org golang:$(GO_VERSION)
APP_NAME := kuby
PLATFORMS := darwin linux
ARCHS := amd64

help: ## print this help
	@grep -E '^[a-zA-Z._-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## compile for your platform in a docker container
	$(DOCKER_CMD) make compile-command

release: ## compile for multiple platforms in a docker container
	$(DOCKER_CMD) make release-command checksums

go-vendor: ## get dependencies
	go mod vendor

release-command: go-vendor ## cross compile command which can also be used on host without docker and also creates checksums
	@for GOOS in ${PLATFORMS}; do \
	  for GOARCH in ${ARCHS}; do \
	  	export GOOS=$$GOOS; \
		export GOARCH=$$GOARCH; \
	    go build -v -o ./build/${APP_NAME}-$$GOOS-$$GOARCH; \
	  done \
	done

compile-command: go-vendor ## compile command which can also be used on host without docker
	go build -v -o ./build/${APP_NAME}

checksums:
	cd ./build/ && sha256sum * > SHA256SUMS

clean:
	rm -r build/*
