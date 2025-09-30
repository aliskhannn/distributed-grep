# Build CLI binary
build:
	go build -o bin/distributed-grep ./cmd/distributed-grep

# Run example (requires servers running on 8081,8082)
run-example:
	go run examples/simple/main.go

# Run all unit tests
test:
	go test -v ./...

# Run only integration test comparing with system grep
integration-test:
	go test ./tests -v -run ^TestDistributedGrepVsGrep$$