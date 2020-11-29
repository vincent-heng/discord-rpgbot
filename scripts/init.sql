CREATE DATABASE rpg;
\connect rpg
CREATE TABLE "characters" (
    "discord_id" bigint PRIMARY KEY,
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

CREATE TABLE "monsters" (
    "monster_id" SERIAL PRIMARY KEY,
    "monster_name" text NOT NULL,
    "experience" integer DEFAULT '0' NOT NULL,
    "strength" integer DEFAULT '1' NOT NULL,
    "agility" integer DEFAULT '1' NOT NULL,
    "wisdom" integer DEFAULT '1' NOT NULL,
    "constitution" integer DEFAULT '1' NOT NULL,
    "current_hp" integer DEFAULT '1' NOT NULL
) WITH (oids = false);
CREATE INDEX monster_queue_current_hp
ON monster_queue (current_hp);

CREATE TABLE "battle_participation" (
  "battle_participation_id" SERIAL PRIMARY KEY,
  "monster_queue_id" integer,
  "character_discord_id" bigint
) WITH (oids = false);

