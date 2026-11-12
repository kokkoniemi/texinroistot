package db

import "time"

// Feedback is a message sent from the Info page's feedback form.
type Feedback struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Message   string    `json:"message"`
	// Contact is the sender's optional name or email address ("" when not given).
	Contact string `json:"contact"`
}

type FeedbackRepository interface {
	Create(feedback Feedback) error
	// List returns the newest feedback first, at most limit messages.
	List(limit int) ([]*Feedback, error)
	// Delete removes a message; false when there was no such message.
	Delete(id int64) (bool, error)
}

type feedbackRepo struct{}

// Create implements FeedbackRepository.
func (*feedbackRepo) Create(feedback Feedback) error {
	var contact any
	if feedback.Contact != "" {
		contact = feedback.Contact
	}
	_, err := Execute(
		"INSERT INTO feedback (message, contact) VALUES ($1, $2);",
		feedback.Message,
		contact,
	)
	return err
}

// List implements FeedbackRepository.
func (*feedbackRepo) List(limit int) ([]*Feedback, error) {
	rows, err := Query(
		`SELECT id, created_at, message, COALESCE(contact, '')
		 FROM feedback
		 ORDER BY created_at DESC, id DESC
		 LIMIT $1;`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feedback := []*Feedback{}
	for rows.Next() {
		var f Feedback
		if err = rows.Scan(&f.ID, &f.CreatedAt, &f.Message, &f.Contact); err != nil {
			return nil, err
		}
		feedback = append(feedback, &f)
	}
	return feedback, rows.Err()
}

// Delete implements FeedbackRepository.
func (*feedbackRepo) Delete(id int64) (bool, error) {
	res, err := Execute("DELETE FROM feedback WHERE id = $1;", id)
	if err != nil {
		return false, err
	}
	deleted, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return deleted > 0, nil
}

func NewFeedbackRepository() FeedbackRepository {
	return &feedbackRepo{}
}
