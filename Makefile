.DEFAULT_GOAL := build

fmt:
	go fmt ./...

.PHONY: fmt

test:
	go test ./... -v

.PHONY: test

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out
	go tool cover -func=coverage.out
	rm coverage.out

.PHONY: test-coverage

vet: fmt
	go vet ./...

.PHONY: vet

build: vet
	go build cmd/main.go

.PHONY: build


run: 
	go run cmd/main.go

.PHONY: run