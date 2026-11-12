BEGIN;

ALTER TABLE public.villains_in_stories
	DROP COLUMN ranks;

COMMIT;
