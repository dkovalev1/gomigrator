BIN := gomigrator
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

compose-test-up:
	docker compose -f deployments/docker-compose.yaml up -d

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/gomigrator

install-lint-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
#	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.63.4

lint: install-lint-deps
	golangci-lint run ./...

unit-test:
	go test -race -count 100 github.com/dkovalev1/gomigrator/cmd github.com/dkovalev1/gomigrator/config github.com/dkovalev1/gomigrator/internal github.com/dkovalev1/gomigrator/pkg github.com/dkovalev1/gomigrator/samples/go

test: compose-test-up unit-test integration
	# go test -v -race -count 100 github.com/dkovalev1/gomigrator/cmd github.com/dkovalev1/gomigrator/config github.com/dkovalev1/gomigrator/internal github.com/dkovalev1/gomigrator/pkg github.com/dkovalev1/gomigrator/samples/go

integration:
	go install github.com/onsi/ginkgo/v2/ginkgo
	ginkgo --repeat=10 integration

clean:
	rm -rf $(BIN)

.PHONY: build clean compose-test-up integration
