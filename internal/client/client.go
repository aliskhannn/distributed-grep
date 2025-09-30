package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aliskhannn/distributed-grep/internal/api/handlers/process"
)

// Client represents a distributed-grep client.
// It holds a list of server addresses and a timeout for HTTP requests.
type Client struct {
	Servers []string      // list of server URLs
	Timeout time.Duration // request timeout
}

// New creates a new Client with the given servers and timeout.
// If timeout <= 0, a default of 5 seconds is used.
func New(servers []string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Client{
		Servers: servers,
		Timeout: timeout,
	}
}

// Grep performs a distributed grep operation across the configured servers.
//
// Parameters:
//   - pattern: string to search for
//   - lines: slice of input lines to search
//   - regex: whether to treat pattern as a regular expression
//   - shardsCount: number of shards to split the input lines
//   - quorum: minimum number of successful server responses required per shard
//
// Returns a slice of matched lines in the same order as the input and an error
// if quorum is not reached for any shard.
func (c *Client) Grep(
	pattern string,
	lines []string,
	regex bool,
	shardsCount int,
	quorum int,
) ([]string, error) {
	if shardsCount <= 0 {
		shardsCount = len(c.Servers)
		if shardsCount == 0 {
			shardsCount = 1
		}
	}
	if quorum <= 0 {
		quorum = len(c.Servers)/2 + 1
	}

	shards := shardLines(lines, shardsCount)
	clientHTTP := &http.Client{Timeout: c.Timeout}
	results := make([][]string, len(shards))

	for si, shard := range shards {
		req := process.Request{Pattern: pattern, Lines: shard, Regex: regex}

		type result struct {
			matches []string
			err     error
		}
		resultsCh := make(chan result, len(c.Servers))
		ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)

		var wg sync.WaitGroup
		for _, srv := range c.Servers {
			wg.Add(1)
			go func(s string) {
				defer wg.Done()
				body, _ := json.Marshal(req)
				httpReq, _ := http.NewRequestWithContext(ctx, "POST", strings.TrimRight(s, "/")+"/process", bytes.NewReader(body))
				httpReq.Header.Set("Content-Type", "application/json")

				resp, err := clientHTTP.Do(httpReq)
				if err != nil {
					resultsCh <- result{err: err}
					return
				}
				defer resp.Body.Close()

				var r process.Response
				if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
					resultsCh <- result{err: err}
					return
				}
				if r.Error != "" {
					resultsCh <- result{err: errors.New(r.Error)}
					return
				}
				resultsCh <- result{matches: r.Matches}
			}(srv)
		}

		// Close resultsCh after all goroutines finish.
		go func() {
			wg.Wait()
			close(resultsCh)
		}()

		// Collect successful results for this shard.
		success, allResults := 0, [][]string{}
		for r := range resultsCh {
			if r.err == nil {
				success++
				allResults = append(allResults, r.matches)
			}
		}
		cancel()

		if success < quorum {
			return nil, fmt.Errorf("quorum not reached for shard %d", si)
		}

		// Take the first successful result for this shard.
		results[si] = allResults[0]
	}

	// Merge results, preserving the input order.
	matchSet := make(map[string]int)
	for _, arr := range results {
		for _, m := range arr {
			matchSet[m]++
		}
	}
	var final []string
	for _, l := range lines {
		if matchSet[l] > 0 {
			final = append(final, l)
			matchSet[l]--
		}
	}
	return final, nil
}

// shardLines splits the input lines into n shards in a round-robin fashion.
func shardLines(lines []string, n int) [][]string {
	if n <= 0 {
		n = 1
	}
	shards := make([][]string, n)
	for i, line := range lines {
		idx := i % n
		shards[idx] = append(shards[idx], line)
	}
	return shards
}

// ReadLines reads all lines from an io.Reader and returns them as a slice of strings.
func ReadLines(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
