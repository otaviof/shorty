APP = shorty
BUILD_DIR ?= build
E2E_TEST_DIR ?= test/e2e
DOCKER_IMAGE ?= "otaviof/$(APP)"
VERSION ?= $(shell cat ./version)

.PHONY: default bootstrap build clean test

default: build

dep:
	go get -u github.com/golang/dep/cmd/dep

bootstrap:
	dep ensure -v -vendor-only

build: clean
	go build -v -o $(BUILD_DIR)/$(APP) cmd/$(APP)/*

build-docker:
	docker build --tag $(DOCKER_IMAGE):$(VERSION) .

clean:
	rm -rf $(BUILD_DIR) > /dev/null

clean-vendor:
	rm -rf ./vendor > /dev/null

release: release-go release-docker
	@echo "# Uploaded $(APP) v$(VERSION)!"

release-go:
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
	curl -s -o .ci/codecov.sh https://codecov.io/bash
	bash .ci/codecov.sh -t $(CODECOV_TOKEN)
