package stories

import "testing"

func TestParseStoryHash(t *testing.T) {
	hash, err := parseStoryHash("  abc123  ")
	if err != nil {
		t.Fatalf("expected valid hash, got error: %v", err)
	}
	if hash != "abc123" {
		t.Fatalf("expected trimmed hash abc123, got %q", hash)
	}
}

func TestParseStoryHashRejectsEmpty(t *testing.T) {
	_, err := parseStoryHash("   ")
	if err == nil {
		t.Fatalf("expected error for empty story hash")
	}
}
