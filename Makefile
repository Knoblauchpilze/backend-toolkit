
GIT_COMMIT_HASH=$(shell git rev-parse --short HEAD)
GOLANGCI_LINT_VERSION ?= v2.12.2

generate-mocks:
	go generate ./...

publish-release:
	./scripts/create_release.sh

tests:
	go test ./...

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} run ./...

fix-lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} run --fix ./...