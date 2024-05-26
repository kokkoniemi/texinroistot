-- This script documents the database structure. It can also be used to create the tables in an empty database 

-- AUTHORS:

-- Table Definition
CREATE TABLE "public"."authors" (
	    "id" int8 NOT NULL,
	    "first_name" varchar,
	    "last_name" varchar,
	    "is_writer" bool,
	    "is_drawer" bool,
	    "is_inventor" bool,
	    "version" int8 NOT NULL,
	    "hash" varchar NOT NULL,
	    PRIMARY KEY ("id")
);

-- Column Comment
COMMENT ON COLUMN "public"."authors"."hash" IS 'consistent identifier between different versions';


DROP TYPE IF EXISTS "public"."author_type";
CREATE TYPE "public"."author_type" AS ENUM ('writer', 'drawer', 'inventor');

-- Table Definition
CREATE TABLE "public"."authors_in_stories" (
	    "id" int8 NOT NULL,
	    "story" int8,
	    "author" int8,
	    "type" "public"."author_type",
	    PRIMARY KEY ("id")
);

-- Comments
COMMENT ON TABLE "public"."authors" IS 'An author of a story. Can be writer, drawer, or story inventor';


-- STORIES:

-- Table Definition
CREATE TABLE "public"."stories" (
	    "id" int8 NOT NULL,
	    "order_num" int8,
	    "version" int8 NOT NULL,
	    "hash" varchar NOT NULL,
	    PRIMARY KEY ("id")
);

-- Column Comment
COMMENT ON COLUMN "public"."stories"."hash" IS 'consistent identifier between different versions';

-- Table Definition
CREATE TABLE "public"."stories_in_publications" (
	    "id" int8 NOT NULL,
	    "story" int8 NOT NULL,
	    "publication" int8 NOT NULL,
	    "title" varchar NOT NULL,
	    PRIMARY KEY ("id")
);

DROP TYPE IF EXISTS "public"."publication_type";
CREATE TYPE "public"."publication_type" AS ENUM ('perus', 'maxi', 'suur', 'muu_erikois', 'kronikka', 'kirjasto', 'italia_perus', 'italia_erikois');

-- Table Definition
CREATE TABLE "public"."publications" (
	    "id" int8 NOT NULL,
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
	    "id" int8 NOT NULL,
	    "created_at" timestamptz NOT NULL DEFAULT now(),
	    "hash" varchar NOT NULL,
	    "is_admin" bool NOT NULL DEFAULT false,
	    PRIMARY KEY ("id")
);


-- VERSIONS:

-- Table Definition
CREATE TABLE "public"."versions" (
	    "id" int8 NOT NULL,
	    "created_at" timestamptz NOT NULL DEFAULT now(),
	    "is_active" bool NOT NULL DEFAULT false,
	    PRIMARY KEY ("id")
);

-- Comments
COMMENT ON TABLE "public"."versions" IS 'Every row in database is related to certain version';


-- VILLAINS:

-- Table Definition
CREATE TABLE "public"."villains" (
	    "id" int8 NOT NULL,
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
	    "id" int8 NOT NULL,
	    "villain" int8 NOT NULL,
	    "story" int8 NOT NULL,
	    "hash" varchar NOT NULL,
	    "nicknames" _varchar,
	    "aliases" _varchar,
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

CREATE UNIQUE INDEX users_hash_key ON public.users USING btree (hash)

