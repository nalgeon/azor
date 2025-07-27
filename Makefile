all: vet lint test

lint:
	@golangci-lint run ./...
	@echo "✓ lint"

vet:
	@go vet ./...
	@echo "✓ vet"

test:
	@go test -race ./...
	@echo "✓ test"
