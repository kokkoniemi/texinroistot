BEGIN;

-- The name a villain goes by in a single story. A villain may be known by
-- one name in an early story and turn out to have another later (e.g. "Kid
-- Rodelo" is revealed to be Philip Rainey). Nullable so the previous release
-- can keep inserting rows without knowing about the columns; empty means the
-- story gives no name of its own and the villain's name applies.
ALTER TABLE public.villains_in_stories
	ADD COLUMN first_names _varchar,
	ADD COLUMN last_name varchar;

COMMIT;
