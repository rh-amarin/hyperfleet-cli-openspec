BINARY     := bin/hf
MODULE     := github.com/rh-amarin/hyperfleet-cli
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -ldflags "-X $(MODULE)/internal/version.Version=$(VERSION) \
                         -X $(MODULE)/internal/version.BuildTime=$(BUILD_TIME)"

.PHONY: build test vet clean lint completions

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

## lint: run static analysis
lint:
	go vet ./...

## completions: generate shell completion scripts into completions/
completions: bin/hf
	@mkdir -p completions
	./bin/hf completion bash  > completions/hf.bash
	./bin/hf completion zsh   > completions/_hf
	./bin/hf completion fish  > completions/hf.fish
