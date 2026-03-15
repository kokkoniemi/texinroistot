package admin

import "testing"

func TestParseVersionID(t *testing.T) {
	id, err := parseVersionID(" 42 ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 42 {
		t.Fatalf("expected 42, got %d", id)
	}
}

func TestParseVersionIDRejectsInvalidValue(t *testing.T) {
	if _, err := parseVersionID("abc"); err == nil {
		t.Fatalf("expected error for non-numeric value")
	}
}

func TestParseVersionIDRejectsZeroOrNegative(t *testing.T) {
	if _, err := parseVersionID("0"); err == nil {
		t.Fatalf("expected error for zero")
	}
	if _, err := parseVersionID("-5"); err == nil {
		t.Fatalf("expected error for negative")
	}
}
