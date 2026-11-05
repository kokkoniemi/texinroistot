BEGIN;

ALTER TABLE public.stories_in_publications
	DROP COLUMN is_republication;

COMMIT;
