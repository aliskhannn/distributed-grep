package grep

import (
	"regexp"
	"runtime"
	"strings"
	"sync"
)

// Grep holds configuration for matching operations.
// Currently, it only stores the number of worker goroutines.
type Grep struct {
	nWorkers int // number of concurrent workers to use
}

// New creates a new Grep processor.
//
// Parameters:
//   - nWorkers: number of worker goroutines for parallel processing.
//     If <= 0, defaults to the number of available CPU cores.
func New(nWorkers int) *Grep {
	if nWorkers <= 0 {
		nWorkers = runtime.GOMAXPROCS(0)
	}
	return &Grep{nWorkers: nWorkers}
}

// Match searches for a pattern in the provided lines.
//
// Parameters:
//   - lines: the input lines to search
//   - pattern: the substring or regex pattern to match
//   - regex: if true, treat pattern as a regular expression
//
// Returns a slice of matched lines and an error if regex compilation fails.
func (g *Grep) Match(lines []string, pattern string, regex bool) ([]string, error) {
	in := make(chan int, len(lines))
	out := make(chan string, len(lines))

	var wg sync.WaitGroup
	wg.Add(g.nWorkers)

	var re *regexp.Regexp
	var err error
	if regex {
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i < g.nWorkers; i++ {
		go func() {
			defer wg.Done()
			for idx := range in {
				line := lines[idx]
				match := false
				if regex {
					if re.MatchString(line) {
						match = true
					}
				} else {
					if strings.Contains(line, pattern) {
						match = true
					}
				}
				if match {
					out <- line
				}
			}
		}()
	}

	for i := range lines {
		in <- i
	}
	close(in)

	go func() {
		wg.Wait()
		close(out)
	}()

	matches := make([]string, 0)
	for m := range out {
		matches = append(matches, m)
	}

	return matches, nil
}
