package main

import "gorm.io/gorm"

// Monster represents an opponent to the characters group
type Monster struct {
	gorm.Model
	MonsterName  string
	Experience   int
	Strength     int
	Agility      int
	Wisdom       int
	Constitution int
	CurrentHp    int
}
