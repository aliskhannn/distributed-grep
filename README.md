# distributed-grep

A distributed CLI utility inspired by the classic `grep`.
It supports parallel string or regex matching across multiple servers with **quorum-based consensus** and concurrent processing.

---

## Features

* Distributed execution across multiple servers
* Quorum-based consensus (`N/2 + 1` by default)
* Parallel processing using goroutines
* Supports substring and regex search
* Clean architecture:

    * `pkg/grep` – core grep logic
    * `internal/api/handlers/process` – HTTP handler
    * `internal/api/server` – server runner
    * `internal/client` – distributed client
* CLI interface (`--mode server` / `--mode client`)
* Examples and tests included

---

## Project Structure

```
distributed-grep/
├── cmd/distributed-grep/          # CLI entrypoint
│   └── main.go
├── internal/
│   ├── client/                    # distributed client
│   └── api/
│       ├── handlers/process/      # HTTP handler
│       └── server/                # server runner
├── pkg/grep/                      # core grep matching logic
├── examples/                      # usage examples
│   └── simple/
├── tests/                         # integration/unit tests
├── Makefile
├── README.md
└── go.mod
```

---

## Installation

```bash
git clone https://github.com/aliskhannn/distributed-grep.git
cd distributed-grep
go build -o bin/distributed-grep ./cmd/distributed-grep
```

---

## Usage

### Start servers

Run multiple servers on different ports:

```bash
bin/distributed-grep --mode server --port 8081
bin/distributed-grep --mode server --port 8082
```

Each server exposes an HTTP API at `/process`.

---

### Run client

Pipe input into the client and search for a substring:

```bash
echo -e "apple\nbanana\napple pie\norange\n" | \
  bin/distributed-grep \
    --mode client \
    --servers http://localhost:8081,http://localhost:8082 \
    --pattern apple
```

Output:

```
apple
apple pie
```

---

### Regex mode

```bash
cat data.txt | \
  bin/distributed-grep \
    --mode client \
    --servers http://localhost:8081,http://localhost:8082 \
    --pattern '^a.*e$' \
    --regex
```

---

## Makefile

```makefile
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
```

---

## Quorum

* Each client request is **sharded** and distributed to all servers.
* A shard is accepted once a **quorum** of servers (majority by default) returns a result.
* This ensures resilience against server failures.

Example:

* 3 servers → quorum = 2
* 5 servers → quorum = 3

---

## Testing

Run integration tests (uses `httptest.Server`):

```bash
go test ./tests -v
```

---

## Example with `examples/simple`

```bash
make build
make run-example
```

Expected output:

```
apple
apple pie
```