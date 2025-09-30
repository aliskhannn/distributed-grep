package tests

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.com/aliskhannn/distributed-grep/internal/api/handlers/process"
	"github.com/aliskhannn/distributed-grep/internal/client"
	"github.com/aliskhannn/distributed-grep/pkg/grep"
)

func makeTestServer() *httptest.Server {
	g := grep.New(2)
	h := process.New(g)
	return httptest.NewServer(http.HandlerFunc(h.Process))
}

func TestDistributedGrepVsGrep(t *testing.T) {
	s1 := makeTestServer()
	s2 := makeTestServer()
	defer s1.Close()
	defer s2.Close()

	servers := []string{s1.URL, s2.URL}

	input := "apple\nbanana\napple pie\norange\n"
	lines := strings.Split(input, "\n")
	lines = lines[:len(lines)-1]

	cl := client.New(servers, 2*time.Second)
	results, err := cl.Grep("apple", lines, false, 2, 1)
	if err != nil {
		t.Fatalf("distributed-grep error: %v", err)
	}

	cmd := exec.Command("grep", "apple")
	cmd.Stdin = strings.NewReader(input)
	expectedBytes, err := cmd.Output()
	if err != nil {
		t.Fatalf("system grep error: %v", err)
	}

	expected := strings.Split(strings.TrimSpace(string(expectedBytes)), "\n")

	if len(results) != len(expected) {
		t.Fatalf("number of matches differ: got %d, want %d", len(results), len(expected))
	}
	for i := range results {
		if results[i] != expected[i] {
			t.Fatalf("match differs at index %d: got %q, want %q", i, results[i], expected[i])
		}
	}
}
