# Optionally set these args as environment vars in the shell. You
# could also pass them as parameters of `make`.
# For example: make build CMD=console
CMD?=wt-tactics
FLAGS?=-v
CLEANUP?=

export GOPRIVATE=github.com/wt-tools

# Requires GNU grep
APP:=$(shell grep -Po '^module\s+\K.*' go.mod)

default: lint test

-include doc.mk

build:
	go build $(FLAGS) $(APP)/cmd/$(CMD)

build-all:
	$(foreach dir,$(wildcard cmd/*), go build $(FLAGS) $(APP)/$(dir);)

build-race:
	go build $(FLAGS) -race $(APP)/cmd/$(CMD)

lint:
	golangci-lint run -v ./...

test:
	go test $(FLAGS) ./...

test-race:
	go test $(FLAGS) -race ./...

generate:
	go generate $(FLAGS) ./...

tidy:
	go get -u github.com/wt-tools/wtscope@latest
	go mod tidy

clean:
	@echo $(CLEANUP)
	$(foreach f,$(CLEANUP),rm -rf $(f);)

.PHONY: build build-race build-all test test-race tidy lint clean
