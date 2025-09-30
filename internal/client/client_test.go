package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aliskhannn/distributed-grep/internal/api/handlers/process"
	"github.com/aliskhannn/distributed-grep/pkg/grep"
)

func makeTestServer(pattern string) *httptest.Server {
	g := grep.New(2)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req process.Request
		_ = json.NewDecoder(r.Body).Decode(&req)
		resp, _ := g.Match(req.Lines, pattern, req.Regex)
		_ = json.NewEncoder(w).Encode(process.Response{Matches: resp})
	}))
}

func TestClient_Grep_Basic(t *testing.T) {
	s1 := makeTestServer("apple")
	s2 := makeTestServer("apple")
	defer s1.Close()
	defer s2.Close()

	lines := strings.Split("apple\nbanana\napple pie\norange\n", "\n")
	lines = lines[:len(lines)-1]

	cl := New([]string{s1.URL, s2.URL}, 2*time.Second)
	results, err := cl.Grep("apple", lines, false, 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(results))
	}
}

func TestClient_Grep_QuorumFail(t *testing.T) {
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "simulated failure", http.StatusInternalServerError)
	}))
	s2 := makeTestServer("apple")
	defer s1.Close()
	defer s2.Close()

	lines := strings.Split("apple\nbanana\napple pie\norange\n", "\n")
	lines = lines[:len(lines)-1]

	cl := New([]string{s1.URL, s2.URL}, 1*time.Second)
	_, err := cl.Grep("apple", lines, false, 2, 2)
	if err == nil {
		t.Fatalf("expected quorum error, got nil")
	}
}
