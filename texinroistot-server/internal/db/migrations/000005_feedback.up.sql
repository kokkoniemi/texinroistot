BEGIN;

-- Feedback sent from the Info page's feedback form. Not tied to a data
-- version, so it survives imports and version deletions. The sender's contact
-- (a name or an email address) is optional.
CREATE TABLE public.feedback (
	id int8 GENERATED ALWAYS AS IDENTITY,
	created_at timestamptz NOT NULL DEFAULT now(),
	message text NOT NULL,
	contact varchar(200),
	PRIMARY KEY (id),
	CONSTRAINT feedback_message_length CHECK (char_length(message) BETWEEN 1 AND 5000)
);

COMMENT ON TABLE public.feedback IS 'Feedback sent from the Info page';

COMMIT;
