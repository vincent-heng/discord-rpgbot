package db

import (
	"gorm.io/gorm"
)

// BattleParticipation is an association between a player and a monster
type BattleParticipation struct {
	gorm.Model
	MonsterID          uint
	CharacterDiscordID uint
}

func (db *DB) FetchBattleParticipants(monsterID uint) (cs []Character, e error) {
	e = db.Joins("JOIN battle_participation bp ON bp.character_discord_id = character.discord_id").
		Where("monster_id = ?", monsterID).
		Find(&cs).Error
	return
}
