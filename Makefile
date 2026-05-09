BINARY    := bin/hf
MODULE    := github.com/rh-amarin/hyperfleet-cli
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags "-X $(MODULE)/internal/version.Version=$(VERSION)"

.PHONY: build test vet clean

## build: compile the hf binary to bin/hf
build:
	@mkdir -p bin
	go build $(LDFLAGS) -o $(BINARY) .

## test: run all unit tests
test:
	go test ./...

## vet: run go vet static analysis
vet:
	go vet ./...

## clean: remove the compiled binary
clean:
	rm -rf bin/
