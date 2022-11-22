all: help

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w ./

test: fmt
	go test ./... -v -count 1 -parallel 1 -race -coverprofile=coverage.txt -covermode=atomic $(TESTARGS) -timeout 180s

build: fmt
ifeq ($(OS),Windows_NT)
	go build -o  kubetool-cli.exe
else
	CGO_ENABLED=0 go build -o kubetool-cli
endif
	