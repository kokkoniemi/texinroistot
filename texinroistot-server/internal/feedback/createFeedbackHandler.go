package feedback

import (
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/kokkoniemi/texinroistot/internal/db"
)

const (
	maxMessageLength = 5000
	maxContactLength = 200
)

type createFeedbackPayload struct {
	Message string `json:"message"`
	Contact string `json:"contact"`
	// Website is a field hidden from people; a bot filling every field fills
	// it too, and its feedback is dropped.
	Website string `json:"website"`
}

// parseFeedback validates the payload. ok is false for a bot's feedback, which
// is answered as if it were saved.
func parseFeedback(payload createFeedbackPayload) (feedback db.Feedback, ok bool, errorMessage string) {
	if strings.TrimSpace(payload.Website) != "" {
		return db.Feedback{}, false, ""
	}
	message := strings.TrimSpace(payload.Message)
	contact := strings.TrimSpace(payload.Contact)
	if message == "" {
		return db.Feedback{}, false, "message is required"
	}
	if utf8.RuneCountInString(message) > maxMessageLength {
		return db.Feedback{}, false, "message is too long"
	}
	if utf8.RuneCountInString(contact) > maxContactLength {
		return db.Feedback{}, false, "contact is too long"
	}
	return db.Feedback{Message: message, Contact: contact}, true, ""
}

func CreateFeedbackHandler(c *fiber.Ctx) error {
	payload := new(createFeedbackPayload)
	if err := c.BodyParser(payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	feedback, ok, errorMessage := parseFeedback(*payload)
	if errorMessage != "" {
		return c.Status(400).JSON(fiber.Map{"error": errorMessage})
	}
	if !ok {
		return c.Status(201).JSON(fiber.Map{"ok": true})
	}

	if err := db.NewFeedbackRepository().Create(feedback); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save feedback"})
	}
	return c.Status(201).JSON(fiber.Map{"ok": true})
}
