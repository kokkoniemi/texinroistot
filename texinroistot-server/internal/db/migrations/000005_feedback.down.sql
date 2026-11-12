BEGIN;

-- Irreversible for the data: dropping the table deletes all received feedback.
-- Take a backup of public.feedback first if it should be kept.
DROP TABLE IF EXISTS public.feedback;

COMMIT;
