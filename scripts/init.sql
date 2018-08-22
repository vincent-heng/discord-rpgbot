CREATE DATABASE rpg;
\connect rpg
CREATE TABLE "character" (
    "name" text PRIMARY KEY,
    "experience" integer DEFAULT '0' NOT NULL
) WITH (oids = false);
