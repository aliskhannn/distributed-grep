package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliskhannn/distributed-grep/internal/client"
)

func main() {
	// List of servers (should be running beforehand).
	servers := []string{
		"http://localhost:8081",
		"http://localhost:8082",
	}

	// Input data to search.
	input := "apple\nbanana\napple pie\norange\n"
	lines := strings.Split(input, "\n")
	if len(lines[len(lines)-1]) == 0 {
		// Remove empty last line if present.
		lines = lines[:len(lines)-1]
	}

	// Create a new client with a 2-second timeout.
	cl := client.New(servers, 2*time.Second)

	// Perform distributed grep.
	results, err := cl.Grep("apple", lines, false, 2, 1) // pattern="apple", regex=false, 2 shards, quorum=1
	if err != nil {
		panic(err)
	}

	// Print matches to stdout.
	fmt.Println("Matches:")
	for _, line := range results {
		fmt.Println(line)
	}
}
