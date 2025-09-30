package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aliskhannn/distributed-grep/internal/api/handlers/process"
	"github.com/aliskhannn/distributed-grep/internal/api/server"
	"github.com/aliskhannn/distributed-grep/internal/client"
	"github.com/aliskhannn/distributed-grep/pkg/grep"
)

func main() {
	// Command-line flags.
	mode := flag.String("mode", "client", "server or client")
	port := flag.Int("port", 8081, "server port")
	serversFlag := flag.String("servers", "", "comma-separated servers for client")
	pattern := flag.String("pattern", "", "search pattern")
	regex := flag.Bool("regex", false, "interpret pattern as regexp")
	shards := flag.Int("shards", 0, "number of shards")
	quorum := flag.Int("quorum", 0, "quorum size (default majority)")
	timeout := flag.Duration("timeout", 5*time.Second, "request timeout")
	flag.Parse()

	// Server mode.
	if *mode == "server" {
		// Create a grep instance and a process handler.
		g := grep.New(0)
		handler := process.New(g)

		// Run HTTP server.
		server.Run(handler, *port)
		return
	}

	// Client mode.

	// Validate that a search pattern is provided.
	if *pattern == "" {
		_, _ = fmt.Fprintln(os.Stderr, "pattern is required")
		os.Exit(1)
	}

	// Parse server addresses from flag.
	var servers []string
	if *serversFlag != "" {
		for _, s := range strings.Split(*serversFlag, ",") {
			servers = append(servers, strings.TrimSpace(s))
		}
	}
	if len(servers) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "no servers specified")
		os.Exit(1)
	}

	// Read all lines from stdin.
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error reading stdin:", err)
		os.Exit(1)
	}

	// Create client and perform distributed grep.
	cl := client.New(servers, *timeout)
	results, err := cl.Grep(*pattern, lines, *regex, *shards, *quorum)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	// Print matching lines to stdout.
	for _, l := range results {
		fmt.Println(l)
	}
}
