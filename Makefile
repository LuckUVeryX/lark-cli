.PHONY: build clean run test deps

BINARY_NAME=lark
BUILD_DIR=.
export LARK_CONFIG_DIR=.lark

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/lark

clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)

run: build
	./$(BINARY_NAME) $(ARGS)

test:
	go test -v ./...

deps:
	go mod tidy
	go mod download

# Install to go bin
install:
	go install ./cmd/lark

# Install to vault tools/bin
install-local:
	go build -o ../bin/lark ./cmd/lark
