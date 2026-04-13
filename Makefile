.PHONY: build test test-redis vet lint clean

build:
	go build ./...

test:
	go test ./... -count=1

test-redis:
	go test ./internal/limiter/ -run Redis -count=1

vet:
	go vet ./...

lint: vet
	@which golangci-lint > /dev/null 2>&1 || { echo "golangci-lint not installed"; exit 1; }
	golangci-lint run ./...

clean:
	rm -rf bin/ coverage.html

cover:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html in your browser"
