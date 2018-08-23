CREATE DATABASE rpg;
\connect rpg
CREATE TABLE "character" (
    "name" text PRIMARY KEY,
    "class" text NOT NULL,
    "experience" integer DEFAULT '0' NOT NULL,
    "level" integer DEFAULT '1' NOT NULL,
    "strength" integer DEFAULT '1' NOT NULL,
    "agility" integer DEFAULT '1' NOT NULL,
    "wisdom" integer DEFAULT '1' NOT NULL,
    "constitution" integer DEFAULT '1' NOT NULL,
    "skill_points" integer DEFAULT '0' NOT NULL,
    "current_hp" integer DEFAULT '1' NOT NULL,
    "stamina" integer  DEFAULT '100' NOT NULL
) WITH (oids = false);

CREATE TABLE "monster_queue" (
    "monster_queue_id" SERIAL PRIMARY KEY,
    "monster_name" text NOT NULL,
    "experience" integer DEFAULT '0' NOT NULL,
    "strength" integer DEFAULT '1' NOT NULL,
    "agility" integer DEFAULT '1' NOT NULL,
    "wisdom" integer DEFAULT '1' NOT NULL,
    "constitution" integer DEFAULT '1' NOT NULL,
    "current_hp" integer DEFAULT '1' NOT NULL
) WITH (oids = false);
