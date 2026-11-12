package admin

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestDeleteFeedbackHandler_RejectsInvalidID(t *testing.T) {
	app := fiber.New()
	app.Delete("/api/admin/feedback/:feedbackID", DeleteFeedbackHandler)

	for _, id := range []string{"abc", "0", "-3"} {
		res, err := app.Test(httptest.NewRequest("DELETE", "/api/admin/feedback/"+id, nil))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if res.StatusCode != 400 {
			t.Fatalf("expected 400 for id %q, got %d", id, res.StatusCode)
		}
	}
}
