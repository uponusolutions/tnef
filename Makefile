
.PHONY: all build test race cover

.DEFAULT_GOAL=test

all: cover lint

build:
	go build

test:
	go test

race:
	go test -race

cover:
	go test -race ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

fumpt-bin:
	@which gofumpt || go install mvdan.cc/gofumpt@latest

fumpt: fumpt-bin
	gofumpt -l -w .

lint: fumpt
	@golangci-lint run
