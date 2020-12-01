package main

import (
	"errors"
	"io/ioutil"
	"os"
	"strconv"

	"gorm.io/gorm"
)

func fetchCharacters() (string, error) {
	characters := []Character{}
	result := db.Select("ID", "level").Find(&characters)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", result.Error
	}

	charactersString := ""
	for _, character := range characters {
		charactersString += discordIDToText(character.ID) + " (niv. " + strconv.Itoa(character.Level) + ") "
	}
	return charactersString, nil
}

func fetchCharacterInfo(tx *gorm.DB, userID uint) (Character, error) {
	var characterInfo Character
	result := tx.First(&characterInfo, userID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return Character{}, result.Error
	}

	return characterInfo, nil
}

func fetchMonsterInfo(tx *gorm.DB) (Monster, error) {
	var monsterInfo Monster
	if e := tx.Where("current_hp > 0").First(&monsterInfo).Error; e != nil {
		return Monster{}, e
	}

	return monsterInfo, nil
}

func fetchMonsterWithParticipants(tx *gorm.DB) (Monster, error) {
	var monsterInfo Monster
	if e := tx.Joins("Character").Where("current_hp > 0").First(&monsterInfo).Error; e != nil {
		return Monster{}, e
	}

	return monsterInfo, nil
}

func createCharacter(discordID uint) error {
	characterToCreate := getDefaultCharacter()
	characterToCreate.ID = discordID

	result := db.Create(&characterToCreate)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	return nil
}

func setAdventureChannel(channelID string) error {
	fileName := "current_channel.txt"
	if !fileExists(fileName) {
		file, err := os.Create(fileName)
		if err != nil {
			return err
		}
		file.Close()
	} else {
		err := os.Truncate(fileName, 0)
		if err != nil {
			return err
		}
	}

	err := ioutil.WriteFile(fileName, []byte(channelID), 0666)
	if err != nil {
		return err
	}

	return nil
}

func spawnMonster(monsterToSpawn Monster) error {
	result := db.Create(&monsterToSpawn)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	return nil
}

func upStats(statsToUp string, userID uint, amount int) error {

	db.Transaction(func(tx *gorm.DB) error {
		character, err := fetchCharacterInfo(tx, userID)
		if err != nil {
			return err
		}
		if amount > character.SkillPoints {
			return errors.New("Not enough skill points")
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
			return errors.New("Wrong stat")
		}
		character.SkillPoints = character.SkillPoints - amount
		tx.Save(&character)
		return nil
	})

	return nil

}
