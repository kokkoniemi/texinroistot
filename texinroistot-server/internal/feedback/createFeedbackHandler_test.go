package feedback

import (
	"strings"
	"testing"
)

func TestParseFeedback_TrimsAndKeepsContact(t *testing.T) {
	feedback, ok, errorMessage := parseFeedback(createFeedbackPayload{
		Message: "  Hyvä sivusto!  ",
		Contact: " Jouni ",
	})
	if !ok || errorMessage != "" {
		t.Fatalf("expected valid feedback, got ok=%v error=%q", ok, errorMessage)
	}
	if feedback.Message != "Hyvä sivusto!" || feedback.Contact != "Jouni" {
		t.Fatalf("expected trimmed fields, got %+v", feedback)
	}
}

func TestParseFeedback_RejectsEmptyAndTooLong(t *testing.T) {
	cases := []createFeedbackPayload{
		{Message: "   "},
		{Message: strings.Repeat("ä", maxMessageLength+1)},
		{Message: "ok", Contact: strings.Repeat("a", maxContactLength+1)},
	}
	for _, payload := range cases {
		if _, ok, errorMessage := parseFeedback(payload); ok || errorMessage == "" {
			t.Fatalf("expected an error for %d-char message, got ok=%v error=%q", len(payload.Message), ok, errorMessage)
		}
	}
	if _, ok, errorMessage := parseFeedback(createFeedbackPayload{Message: strings.Repeat("ä", maxMessageLength)}); !ok || errorMessage != "" {
		t.Fatalf("expected a message of the maximum length to pass, got ok=%v error=%q", ok, errorMessage)
	}
}

func TestParseFeedback_DropsBotFeedbackSilently(t *testing.T) {
	_, ok, errorMessage := parseFeedback(createFeedbackPayload{Message: "spam", Website: "http://spam"})
	if ok || errorMessage != "" {
		t.Fatalf("expected bot feedback to be dropped without an error, got ok=%v error=%q", ok, errorMessage)
	}
}
