package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Configuration filled from configuration file
type Configuration struct {
	DiscordBotKey string
	GameMaster    uint
}

const (
	dbName = "rpg"
)

func initDb() (*gorm.DB, error) {
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
			return nil, fmt.Errorf("Automigrate %+v failed: %w", table, e)
		}
	}
	return db, nil
}

// loadConfiguration loads configuration from json file
func loadConfiguration(configurationFile string) (Configuration, error) {
	configuration := Configuration{}

	file, err := os.Open(configurationFile)
	if err != nil {
		return configuration, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	return configuration, err
}

func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func getChannelID() (string, error) {
	return readFile("current_channel.txt")
}

func readFile(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", nil
	}
	scanner := bufio.NewScanner(file)

	success := scanner.Scan()
	if success == false {
		// False on error or EOF. Check error
		err = scanner.Err()
		if err == nil {
			return "", nil
		}
		return "", err
	}

	return scanner.Text(), nil
}

func characterToString(characterInfo Character) string {
	characterString := discordIDToText(characterInfo.ID) + " (" + characterInfo.Class + ") - " + strconv.Itoa(characterInfo.CurrentHp) + " / " + strconv.Itoa(getMaxHP(characterInfo)) + " HP\n"
	characterString = characterString + "Endurance : " + strconv.Itoa(characterInfo.Stamina) + " / 100\n"
	characterString = characterString + "Niveau " + strconv.Itoa(characterInfo.Level) + " (" + strconv.Itoa(characterInfo.Experience) + " XP)\n"
	characterString = characterString + "Force : " + strconv.Itoa(characterInfo.Strength) + "\n"
	characterString = characterString + "Agilité : " + strconv.Itoa(characterInfo.Agility) + "\n"
	characterString = characterString + "Sagesse : " + strconv.Itoa(characterInfo.Wisdom) + "\n"
	characterString = characterString + "Constitution : " + strconv.Itoa(characterInfo.Constitution) + "\n"

	if characterInfo.SkillPoints > 0 {
		plural := ""
		if characterInfo.SkillPoints > 1 {
			plural = "s"
		}
		characterString = characterString + "\nIl vous reste " + strconv.Itoa(characterInfo.SkillPoints) + " point" + plural + " à répartir.\n"
	}

	return characterString
}

func monsterToString(monsterInfo Monster) string {
	return monsterInfo.MonsterName + " - " + strconv.Itoa(monsterInfo.CurrentHp) + " / " + strconv.Itoa(getMaxHPMonster(monsterInfo)) + " HP\n"
}

func getMaxHP(characterInfo Character) int {
	return 10 + characterInfo.Constitution + characterInfo.Level
}

func getMaxHPMonster(monsterInfo Monster) int {
	return 10 + monsterInfo.Constitution
}

func getDefaultCharacter() Character {
	characterToCreate := Character{}
	characterToCreate.Class = "Combattant"
	characterToCreate.Experience = 0
	characterToCreate.Level = 1
	characterToCreate.Strength = 1
	characterToCreate.Agility = 1
	characterToCreate.Wisdom = 1
	characterToCreate.Constitution = 1
	characterToCreate.SkillPoints = 5
	characterToCreate.CurrentHp = getMaxHP(characterToCreate)
	characterToCreate.Stamina = 100
	return characterToCreate
}

func parseLevel(experience int) int {
	floatExperience := float64(experience)
	floatLevel := (-1 + math.Sqrt(1+2*floatExperience/25)) / 2
	roundLevel := int(floatLevel) + 1 // 99 XP -> 0.9 floatLevel -> 1 roundLevel
	return roundLevel
}

func discordIDToText(userID uint) string {
	return "<@" + strconv.FormatUint(uint64(userID), 10) + ">"
}
