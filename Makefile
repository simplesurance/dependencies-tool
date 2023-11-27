default: all

.PHONY: dependencies-tool
dependencies-tool:
	CGO_ENABLED=0 go build -trimpath -o $@  *.go

.PHONY: check
check:
	golangci-lint run

.PHONY: test
test:
	go test -race ./...

.PHONY: all
all: dependencies-tool check test
