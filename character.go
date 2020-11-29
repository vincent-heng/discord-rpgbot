package main

import "gorm.io/gorm"

// Character represents a character in DB, played by a discord user
type Character struct {
	gorm.Model
	Class        string
	Experience   int
	Level        int
	Strength     int
	Agility      int
	Wisdom       int
	Constitution int
	SkillPoints  int
	CurrentHp    int
	Stamina      int
}

// BattleParticipation is an association between a player and a monster
type BattleParticipation struct {
	gorm.Model
	MonsterID          uint
	CharacterDiscordID uint
}
