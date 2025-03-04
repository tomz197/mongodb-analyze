
build:
	@echo "Building..."
	@go build -o bin/ ./cmd/...

run:
	@echo "Running..."
	@go run ./cmd/...

test:
	@echo "Testing..."
	@go test -v ./...
