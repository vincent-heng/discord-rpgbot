package db

import (
	"strconv"

	"gorm.io/gorm"
)

type Monster struct {
	gorm.Model
	Name         string
	Experience   int
	Strength     int
	Agility      int
	Wisdom       int
	Constitution int
	CurrentHp    int
	Participants []*Character `gorm:"many2many:battle_participations;"`
}

func (m Monster) String() string {
	return m.Name + " - " + strconv.Itoa(m.CurrentHp) + " / " + strconv.Itoa(m.GetMaxHP()) + " HP\n"
}

func (m Monster) GetMaxHP() int {
	return 10 + m.Constitution
}

func (db *DB) FetchMonsterInfo() (m Monster, e error) {
	e = db.Where("current_hp > 0").First(&m).Error
	return
}

func (db *DB) FetchMonsterWithParticipants() (m Monster, e error) {
	e = db.Where("current_hp > 0").First(&m).Error
	return
}

func (db *DB) SpawnMonster(m Monster) error {
	return db.Create(&m).Error
}
