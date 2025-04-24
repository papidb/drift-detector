.DEFAULT_GOAL := build

fmt:
	go fmt ./...

.PHONY: fmt

test:
	go test ./...

.PHONY: test

test-verbose:
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

generate-example-data: 
	aws ec2 describe-instances --region us-west-2 > sample-data/ec2-instances.json

.PHONY: generate-example-data


tf-apply: 
	cd infrastructure && terraform apply -auto-approve
.PHONY: tf-apply

tf-destroy: 
	cd infrastructure  &&	terraform destroy -auto-approve
.PHONY: tf-destroy

drift: 
	cd scripts/ && chmod +x drift.sh && ./drift.sh

.PHONY: drift

destroy-all-instances:
	cd scripts/ && chmod +x destroy.sh && ./destroy.sh
.PHONY: destroy-all-instances
