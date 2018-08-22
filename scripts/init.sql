CREATE DATABASE rpg;
\connect rpg
CREATE TABLE "character" (
    "name" text PRIMARY KEY,
    "experience" integer DEFAULT '0' NOT NULL
) WITH (oids = false);

CREATE TABLE "monster_queue" (
    "monster_queue_id" SERIAL PRIMARY KEY,
    "monster_name" text NOT NULL,
    "current_hp" integer DEFAULT '1' NOT NULL,
    "max_hp" integer DEFAULT '1' NOT NULL,
    "experience" integer DEFAULT '0' NOT NULL
) WITH (oids = false);
