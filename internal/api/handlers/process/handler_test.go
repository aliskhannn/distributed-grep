package process

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aliskhannn/distributed-grep/pkg/grep"
)

func TestHandler_Process_Substring(t *testing.T) {
	g := grep.New(2)
	h := New(g)

	reqBody := Request{
		Pattern: "foo",
		Lines:   []string{"foo", "bar", "foobar"},
		Regex:   false,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/process", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Process(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(resp.Matches))
	}
}
