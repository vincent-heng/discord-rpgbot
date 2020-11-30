package db

import (
	"errors"
	"strconv"

	"gorm.io/gorm"

	"github.com/vincent-heng/discord-airpgbot/bot/util"
)

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

func (c Character) String() string {
	str := util.DiscordIDToText(c.ID) + " (" + c.Class + ") - " +
		strconv.Itoa(c.CurrentHp) + " / " + strconv.Itoa(c.GetMaxHP()) + " HP\n" +
		"Endurance : " + strconv.Itoa(c.Stamina) + " / 100\n" +
		"Niveau " + strconv.Itoa(c.Level) + " (" + strconv.Itoa(c.Experience) + " XP)\n" +
		"Force : " + strconv.Itoa(c.Strength) + "\n" +
		"Agilité : " + strconv.Itoa(c.Agility) + "\n" +
		"Sagesse : " + strconv.Itoa(c.Wisdom) + "\n" +
		"Constitution : " + strconv.Itoa(c.Constitution) + "\n"

	if c.SkillPoints > 0 {
		plural := ""
		if c.SkillPoints > 1 {
			plural = "s"
		}
		str += "\nIl vous reste " + strconv.Itoa(c.SkillPoints) + " point" + plural + " à répartir.\n"
	}

	return str
}

func (c Character) GetMaxHP() int {
	return 10 + c.Constitution + c.Level
}

func NewCharacter() Character {
	c := Character{
		Class:        "Combattant",
		Experience:   0,
		Level:        1,
		Strength:     1,
		Agility:      1,
		Wisdom:       1,
		Constitution: 1,
		SkillPoints:  5,
		Stamina:      100,
	}
	c.CurrentHp = c.GetMaxHP()
	return c
}

func (db *DB) FetchCharacters() (string, error) {
	characters := []Character{}
	result := db.Select("ID", "level").Find(&characters)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", result.Error
	}

	charactersString := ""
	for i := range characters {
		character := &characters[i]
		charactersString += util.DiscordIDToText(character.ID) + " (niv. " + strconv.Itoa(character.Level) + ") "
	}
	return charactersString, nil
}

func (db *DB) FetchCharacterInfo(userID uint) (c Character, e error) {
	e = db.First(&c, userID).Error
	return
}

func (db *DB) CreateCharacter(discordID uint) error {
	characterToCreate := NewCharacter()
	characterToCreate.ID = discordID

	result := db.Create(&characterToCreate)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	return nil
}

func (db *DB) UpStats(statsToUp string, userID uint, amount int) error {
	tx := db.Begin()
	defer tx.Rollback()

	character, err := tx.FetchCharacterInfo(userID)
	if err != nil {
		return err
	}

	if amount > character.SkillPoints {
		return errNotEnoughSkillPoints
	}

	switch statsToUp {
	case "strength":
		character.Strength = character.Strength + amount
	case "agility":
		character.Agility = character.Agility + amount
	case "wisdom":
		character.Wisdom = character.Wisdom + amount
	case "constitution":
		character.Constitution = character.Constitution + amount
	default:
		return errWrongStat
	}

	character.SkillPoints = character.SkillPoints - amount

	if err := tx.Save(&character).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}
