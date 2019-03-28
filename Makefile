# application name, used in pkg, cmd and other places
APP = shorty
# build directory
BUILD_DIR ?= build
# temporary sqlite file for "go run"
TEMP_DATABASE_FILE ?= .ci/shorty.sqlite
# directory containing end-to-end tests
E2E_TEST_DIR ?= test/e2e
# docker image name
DOCKER_IMAGE ?= "otaviof/$(APP)"
# project version, and also docker image tag
VERSION ?= $(shell cat ./version)

.PHONY: default bootstrap build clean test

default: build

dep:
	go get -u github.com/golang/dep/cmd/dep

bootstrap:
	dep ensure -v -vendor-only

run:
	go run cmd/shorty/shorty.go --database-file $(TEMP_DATABASE_FILE)

build: clean
	go build -v -o $(BUILD_DIR)/$(APP) cmd/$(APP)/*

build-docker:
	docker build --tag $(DOCKER_IMAGE):$(VERSION) .

clean:
	rm -rf $(BUILD_DIR) > /dev/null

clean-vendor:
	rm -rf ./vendor > /dev/null

release:
	git tag $(VERSION)
	git push origin $(VERSION)
	goreleaser --rm-dist

release-docker: build-docker
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest

snapshot:
	goreleaser --rm-dist --snapshot

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -cover -v pkg/$(APP)/*

integration:
	go test -v $(E2E_TEST_DIR)/*

codecov:
	mkdir .ci || true
	curl -s -o .ci/codecov.sh https://codecov.io/bash
	bash .ci/codecov.sh -t $(CODECOV_TOKEN)
