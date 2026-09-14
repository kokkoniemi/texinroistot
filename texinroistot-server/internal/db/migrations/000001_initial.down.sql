BEGIN;

DROP TABLE public.villains_in_stories;
DROP TABLE public.stories_in_publications;
DROP TABLE public.authors_in_stories;
DROP TABLE public.villains;
DROP TABLE public.publications;
DROP TABLE public.stories;
DROP TABLE public.authors;
DROP TABLE public.users;
DROP TABLE public.versions;

DROP TYPE public.publication_type;
DROP TYPE public.author_type;

COMMIT;
