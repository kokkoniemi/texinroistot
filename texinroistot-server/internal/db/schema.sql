-- This script documents the database structure. It can also be used to create the tables in an empty database 

-- AUTHORS:

-- Table Definition
CREATE TABLE "public"."authors" (
	    "id" int8 GENERATED ALWAYS AS IDENTITY,
	    "first_name" varchar,
	    "last_name" varchar,
	    "is_writer" bool,
	    "is_drawer" bool,
	    "is_translator" bool,
	    "version" int8 NOT NULL,
	    "hash" varchar NOT NULL,
	    PRIMARY KEY ("id")
);

-- Column Comment
COMMENT ON COLUMN "public"."authors"."hash" IS 'consistent identifier between different versions';


DROP TYPE IF EXISTS "public"."author_type";
CREATE TYPE "public"."author_type" AS ENUM ('writer', 'drawer', 'translator');

-- Table Definition
CREATE TABLE "public"."authors_in_stories" (
	    "id" int8 GENERATED ALWAYS AS IDENTITY,
	    "story" int8,
	    "author" int8,
	    "type" "public"."author_type",
	    "details" varchar,
	    PRIMARY KEY ("id")
);

-- Comments
COMMENT ON TABLE "public"."authors" IS 'An author of a story. Can be writer, drawer, or story translator';


-- STORIES:

-- Table Definition
CREATE TABLE "public"."stories" (
	    "id" int8 GENERATED ALWAYS AS IDENTITY,
	    "order_num" int8,
	    "version" int8 NOT NULL,
	    "hash" varchar NOT NULL,
	    PRIMARY KEY ("id")
);

-- Column Comment
COMMENT ON COLUMN "public"."stories"."hash" IS 'consistent identifier between different versions';

-- Table Definition
CREATE TABLE "public"."stories_in_publications" (
	    "id" int8 GENERATED ALWAYS AS IDENTITY,
	    "story" int8 NOT NULL,
	    "publication" int8 NOT NULL,
	    "title" varchar NOT NULL,
	    PRIMARY KEY ("id")
);

DROP TYPE IF EXISTS "public"."publication_type";
CREATE TYPE "public"."publication_type" AS ENUM (
	'perus',
	'maxi',
	'suur',
	'muu_erikois',
	'kronikka',
	'kirjasto',
	'italia_perus',
	'italia_erikois',
	'italia_serie_extra',
	'italia_texone',
	'italia_mini_texone_maxi_tex',
	'italia_almanacco_del_west',
	'italia_color_tex',
	'italia_tex_romanzi_a_fumetti',
	'italia_tex_magazine'
);

-- Table Definition
CREATE TABLE "public"."publications" (
	    "id" int8 GENERATED ALWAYS AS IDENTITY,
	    "type" "public"."publication_type",
	    "year" int8,
	    "issue" varchar NOT NULL,
	    "version" int8 NOT NULL,
	    "hash" varchar NOT NULL,
	    PRIMARY KEY ("id")
);

-- Column Comment
COMMENT ON COLUMN "public"."publications"."hash" IS 'consistent identifier between different versions';


-- USERS:

-- Table Definition
CREATE TABLE "public"."users" (
	    "id" int8 GENERATED ALWAYS AS IDENTITY,
	    "created_at" timestamptz NOT NULL DEFAULT now(),
	    "hash" varchar NOT NULL,
	    "is_admin" bool NOT NULL DEFAULT false,
	    PRIMARY KEY ("id")
);


-- VERSIONS:

-- Table Definition
CREATE TABLE "public"."versions" (
	    "id" int8 GENERATED ALWAYS AS IDENTITY,
	    "created_at" timestamptz NOT NULL DEFAULT now(),
	    "is_active" bool NOT NULL DEFAULT false,
	    PRIMARY KEY ("id")
);

-- Comments
COMMENT ON TABLE "public"."versions" IS 'Every row in database is related to certain version';


-- VILLAINS:

-- Table Definition
CREATE TABLE "public"."villains" (
	    "id" int8 GENERATED ALWAYS AS IDENTITY,
	    "ranks" _varchar,
	    "first_names" _varchar,
	    "last_name" varchar,
	    "version" int8 NOT NULL,
	    "hash" varchar NOT NULL,
	    PRIMARY KEY ("id")
);

-- Column Comment
COMMENT ON COLUMN "public"."villains"."hash" IS 'consistent identifier between different versions';


-- Table Definition
CREATE TABLE "public"."villains_in_stories" (
	    "id" int8 GENERATED ALWAYS AS IDENTITY,
	    "villain" int8 NOT NULL,
	    "story" int8 NOT NULL,
	    "hash" varchar NOT NULL,
	    "nicknames" _varchar,
	    "other_names" _varchar,
	    "code_names" _varchar,
	    "destiny" _varchar,
	    "roles" _varchar,
	    PRIMARY KEY ("id")
);


-- FOREIGN KEY & INDICE DEFINITIONS
ALTER TABLE "public"."authors" ADD FOREIGN KEY ("version") REFERENCES "public"."versions"("id") ON DELETE CASCADE;
ALTER TABLE "public"."authors_in_stories" ADD FOREIGN KEY ("story") REFERENCES "public"."stories"("id") ON DELETE CASCADE;
ALTER TABLE "public"."authors_in_stories" ADD FOREIGN KEY ("author") REFERENCES "public"."authors"("id") ON DELETE CASCADE;
ALTER TABLE "public"."publications" ADD FOREIGN KEY ("version") REFERENCES "public"."versions"("id") ON DELETE CASCADE;
ALTER TABLE "public"."stories" ADD FOREIGN KEY ("version") REFERENCES "public"."versions"("id") ON DELETE CASCADE;
ALTER TABLE "public"."stories_in_publications" ADD FOREIGN KEY ("story") REFERENCES "public"."stories"("id") ON DELETE CASCADE;
ALTER TABLE "public"."stories_in_publications" ADD FOREIGN KEY ("publication") REFERENCES "public"."publications"("id") ON DELETE CASCADE;
ALTER TABLE "public"."villains" ADD FOREIGN KEY ("version") REFERENCES "public"."versions"("id") ON DELETE CASCADE;
ALTER TABLE "public"."villains_in_stories" ADD FOREIGN KEY ("story") REFERENCES "public"."stories"("id") ON DELETE CASCADE;
ALTER TABLE "public"."villains_in_stories" ADD FOREIGN KEY ("villain") REFERENCES "public"."villains"("id") ON DELETE CASCADE;

CREATE UNIQUE INDEX users_hash_key ON public.users USING btree (hash);

-- Query performance indexes for listing and filtering endpoints
CREATE INDEX IF NOT EXISTS idx_villains_version ON public.villains USING btree (version);
CREATE INDEX IF NOT EXISTS idx_stories_version_hash ON public.stories USING btree (version, hash);
CREATE INDEX IF NOT EXISTS idx_villains_in_stories_villain ON public.villains_in_stories USING btree (villain);
CREATE INDEX IF NOT EXISTS idx_villains_in_stories_story ON public.villains_in_stories USING btree (story);
CREATE INDEX IF NOT EXISTS idx_stories_in_publications_story ON public.stories_in_publications USING btree (story);
CREATE INDEX IF NOT EXISTS idx_stories_in_publications_publication ON public.stories_in_publications USING btree (publication);
CREATE INDEX IF NOT EXISTS idx_publications_type ON public.publications USING btree (type);
CREATE INDEX IF NOT EXISTS idx_authors_in_stories_story ON public.authors_in_stories USING btree (story);
