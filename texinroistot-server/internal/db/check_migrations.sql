BEGIN;

DO $$
DECLARE
	version_id bigint;
	author_id bigint;
	story_id bigint;
	publication_id bigint;
	villain_id bigint;
BEGIN
	IF to_regclass('public.users_hash_key') IS NULL
		OR to_regclass('public.idx_villains_version') IS NULL
		OR to_regclass('public.idx_stories_version_hash') IS NULL THEN
		RAISE EXCEPTION 'Required application indexes are missing';
	END IF;

	INSERT INTO public.versions DEFAULT VALUES RETURNING id INTO version_id;
	INSERT INTO public.users (hash) VALUES ('migration-check-user');
	IF NOT EXISTS (SELECT 1 FROM public.users WHERE hash = 'migration-check-user' AND NOT is_admin AND created_at IS NOT NULL) THEN
		RAISE EXCEPTION 'User defaults are incorrect';
	END IF;
	BEGIN
		INSERT INTO public.users (hash) VALUES ('migration-check-user');
		RAISE EXCEPTION 'Duplicate user hashes must be rejected';
	EXCEPTION WHEN unique_violation THEN
		NULL;
	END;

	INSERT INTO public.authors (version, hash) VALUES (version_id, 'migration-check-author') RETURNING id INTO author_id;
	INSERT INTO public.stories (version, hash) VALUES (version_id, 'migration-check-story') RETURNING id INTO story_id;
	INSERT INTO public.publications (version, hash, issue, type)
		VALUES (version_id, 'migration-check-publication', '1', 'perus') RETURNING id INTO publication_id;
	INSERT INTO public.villains (version, hash, ranks)
		VALUES (version_id, 'migration-check-villain', ARRAY['test']) RETURNING id INTO villain_id;
	INSERT INTO public.authors_in_stories (story, author, type) VALUES (story_id, author_id, 'writer');
	INSERT INTO public.stories_in_publications (story, publication, title) VALUES (story_id, publication_id, 'Migration check');
	INSERT INTO public.villains_in_stories (villain, story, hash) VALUES (villain_id, story_id, 'migration-check-link');

	DELETE FROM public.versions WHERE id = version_id;
	IF EXISTS (SELECT 1 FROM public.authors WHERE version = version_id)
		OR EXISTS (SELECT 1 FROM public.stories WHERE version = version_id)
		OR EXISTS (SELECT 1 FROM public.publications WHERE version = version_id)
		OR EXISTS (SELECT 1 FROM public.villains WHERE version = version_id)
		OR EXISTS (SELECT 1 FROM public.authors_in_stories WHERE story = story_id)
		OR EXISTS (SELECT 1 FROM public.stories_in_publications WHERE story = story_id)
		OR EXISTS (SELECT 1 FROM public.villains_in_stories WHERE story = story_id) THEN
		RAISE EXCEPTION 'Version deletion did not cascade through all content tables';
	END IF;
END;
$$;

ROLLBACK;
