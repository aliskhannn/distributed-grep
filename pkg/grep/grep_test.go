package grep

import (
	"testing"
)

func TestGrep_MatchSubstring(t *testing.T) {
	g := New(2)
	lines := []string{"apple", "banana", "apple pie", "orange"}
	matches, err := g.Match(lines, "apple", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"apple", "apple pie"}
	if len(matches) != len(expected) {
		t.Fatalf("expected %d matches, got %d", len(expected), len(matches))
	}
}

func TestGrep_MatchRegex(t *testing.T) {
	g := New(2)
	lines := []string{"cat", "car", "dog", "cart"}
	matches, err := g.Match(lines, `^ca.*`, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"cat", "car", "cart"}
	if len(matches) != len(expected) {
		t.Fatalf("expected %d matches, got %d", len(expected), len(matches))
	}
}

func TestGrep_InvalidRegex(t *testing.T) {
	g := New(2)
	lines := []string{"test"}
	_, err := g.Match(lines, "(*", true)
	if err == nil {
		t.Fatalf("expected error for invalid regex")
	}
}
