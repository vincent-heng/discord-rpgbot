package db

import (
	"errors"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

const (
	dbName = "rpg"
)

func New() (*DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName)

	db, err := gorm.Open(postgres.Open(dbinfo), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	for _, table := range []interface{}{
		&Character{}, &Monster{}, &BattleParticipation{},
	} {
		if e := db.AutoMigrate(table); e != nil {
			return nil, fmt.Errorf("automigrate %+v failed: %w", table, e)
		}
	}
	return &DB{DB: db}, nil
}

func (db *DB) Begin() *DB {
	return &DB{
		DB: db.DB.Begin(),
	}
}

func (db *DB) GetParticipants(characterDiscordID uint) (*Character, *Monster, error) {
	tx := db.Begin()
	defer tx.Rollback()

	attacker, err := tx.FetchCharacterInfo(characterDiscordID)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get character info: %w", err)
	}

	monsterTarget, err := tx.FetchMonsterInfo()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, fmt.Errorf("monster not found: %w", err)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("cannot load monster: %w", err)
	}

	return &attacker, &monsterTarget, nil
}
