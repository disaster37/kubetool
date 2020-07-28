all: help

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w ./

test: fmt
	go test ./... -v -count 1 -parallel 1 -race -coverprofile=coverage.txt -covermode=atomic $(TESTARGS) -timeout 120s

build: fmt
ifeq ($(OS),Windows_NT)
	name :=  kubetools.exe
else
	name := kubetools
endif
	go build -o $(name)