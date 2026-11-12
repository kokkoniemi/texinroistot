BEGIN;

-- The ranks a villain has in a single story. A villain's ranks gather from all
-- of their stories (Quantrill is "Kenraali" in some), but a story shows only
-- the ranks its own row gives. Nullable so the previous release can keep
-- inserting rows without knowing about the column; NULL means the row was
-- imported before the column existed and the villain's ranks apply.
ALTER TABLE public.villains_in_stories
	ADD COLUMN ranks _varchar;

COMMIT;
