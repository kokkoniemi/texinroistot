BEGIN;

-- Marks links created from the Finnish base series re-publication columns
-- (UVuosi, Ualkaen, Upäättyen). Defaults to false so the previous release
-- can keep inserting rows without knowing about the column.
ALTER TABLE public.stories_in_publications
	ADD COLUMN is_republication bool NOT NULL DEFAULT false;

COMMIT;
