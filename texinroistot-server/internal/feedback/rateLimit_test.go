package feedback

import (
	"testing"
	"time"
)

func TestSendLog_LimitsSendsWithinWindow(t *testing.T) {
	log := &sendLog{sends: map[string][]time.Time{}}
	start := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 2; i++ {
		if !log.allow("a", start, 2, time.Minute) {
			t.Fatalf("expected send %d to be allowed", i+1)
		}
	}
	if log.allow("a", start.Add(30*time.Second), 2, time.Minute) {
		t.Fatal("expected the third send within the window to be refused")
	}
	if !log.allow("b", start, 2, time.Minute) {
		t.Fatal("expected another sender to be allowed")
	}
	if !log.allow("a", start.Add(time.Minute), 2, time.Minute) {
		t.Fatal("expected a send after the window to be allowed")
	}
}
