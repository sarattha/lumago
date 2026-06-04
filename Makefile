.PHONY: fmt test run vet tidy

fmt:
	go fmt ./...

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

run:
	go run ./cmd/sandbox
