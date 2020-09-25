APP = shorty
OUTPUT_DIR ?= _output

CMD = ./cmd/$(APP)/...
PKG = ./pkg/...
E2E = ./test/e2e/...

BIN ?= $(OUTPUT_DIR)/$(APP)

GO_FLAGS ?= -v -mod=vendor
GO_TEST_FLAGS ?= -race -cover

ARGS ?=

# temporary sqlite file for "go run"
TEMP_DATABASE_FILE ?= .ci/shorty.sqlite

# docker image name
IMAGE ?= "otaviof/$(APP)"
# project version, and also docker image tag
VERSION ?= $(shell cat ./version)

# codecov authentication token
CODECOV_TOKEN ?=

default: $(BIN)

.PHONY: $(BIN)
$(BIN):
	go build $(GO_FLAGS) -o $(BIN) $(CMD)

vendor:
	go mod vendor

install:
	go install $(GO_FLAGS) $(CMD)

run:
	go run $(GO_FLAGS) $(CMD) $(ARGS)

serve: ARGS = --database-file $(TEMP_DATABASE_FILE)
serve: run

clean:
	rm -rf $(OUTPUT_DIR) > /dev/null

test: test-unit test-e2e

.PHONY: test-unit
test-unit:
	go test $(GO_FLAGS) $(GO_TEST_FLAGS) $(CMD) $(PKG) $(ARGS)

test-e2e:
	go test $(GO_FLAGS) $(GO_TEST_FLAGS) $(E2E) $(ARGS)

codecov:
	mkdir .ci || true
	curl -s -o .ci/codecov.sh https://codecov.io/bash
	bash .ci/codecov.sh -t $(CODECOV_TOKEN)

image:
	docker build --tag $(IMAGE):$(VERSION) .

release:
	git tag $(VERSION)
	git push origin $(VERSION)
	goreleaser --rm-dist

snapshot:
	goreleaser --rm-dist --snapshot
