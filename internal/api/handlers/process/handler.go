package process

import (
	"encoding/json"
	"net/http"
)

// Matcher defines an interface for objects that can match lines against a pattern.
type Matcher interface {
	// Match returns the lines that match the given pattern.
	// If regex is true, pattern is treated as a regular expression.
	Match(lines []string, pattern string, regex bool) ([]string, error)
}

// Handler provides an HTTP handler for processing distributed grep requests.
type Handler struct {
	grep Matcher
}

// New creates a new Handler with the provided Matcher.
func New(grep Matcher) *Handler {
	return &Handler{grep: grep}
}

// Request represents a single grep request payload.
type Request struct {
	Pattern string   `json:"pattern"` // pattern to search for
	Lines   []string `json:"lines"`   // lines to search in
	Regex   bool     `json:"regex"`   // whether to interpret Pattern as regex
}

// Response represents a single grep response payload.
type Response struct {
	Matches []string `json:"matches"` // matched lines
	Error   string   `json:"error"`   // error message, if any
}

// Process handles HTTP requests to perform a grep operation.
// It expects a JSON payload with fields "pattern", "lines", and "regex".
// Returns JSON with either matches or an error.
func (h *Handler) Process(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	matches, err := h.grep.Match(req.Lines, req.Pattern, req.Regex)
	if err != nil {
		_ = json.NewEncoder(w).Encode(&Response{Error: err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(&Response{Matches: matches})
}
