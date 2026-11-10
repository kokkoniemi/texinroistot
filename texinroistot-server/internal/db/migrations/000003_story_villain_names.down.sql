BEGIN;

ALTER TABLE public.villains_in_stories
	DROP COLUMN last_name,
	DROP COLUMN first_names;

COMMIT;
