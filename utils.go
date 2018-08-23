package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// Configuration filled from configuration file
type Configuration struct {
	DiscordBotKey string
	GameMaster    string
}

func initDb() (*sql.DB, error) {
	DB_HOST := os.Getenv("DB_HOST")
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")

	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_USER, DB_PASSWORD, DB_NAME)
	return sql.Open("postgres", dbinfo)
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

func getChannelId() (string, error) {
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
		} else {
			return "", err
		}
	}

	return scanner.Text(), nil
}

func characterToString(characterInfo character) string {
	characterString := characterInfo.name + " (" + characterInfo.class + ") - " + strconv.Itoa(characterInfo.currentHp) + " / " + strconv.Itoa(getMaxHP(characterInfo)) + " HP\n"
	characterString = characterString + "Endurance : " + strconv.Itoa(characterInfo.stamina) + " / 100\n"
	characterString = characterString + "Niveau " + strconv.Itoa(characterInfo.level) + " (" + strconv.Itoa(characterInfo.experience) + " XP)\n"
	characterString = characterString + "Force : " + strconv.Itoa(characterInfo.strength) + "\n"
	characterString = characterString + "Agilité : " + strconv.Itoa(characterInfo.agility) + "\n"
	characterString = characterString + "Sagesse : " + strconv.Itoa(characterInfo.wisdom) + "\n"
	characterString = characterString + "Constitution : " + strconv.Itoa(characterInfo.constitution) + "\n"

	if characterInfo.skillPoints > 0 {
		plural := ""
		if characterInfo.skillPoints > 1 {
			plural = "s"
		}
		characterString = characterString + "\nIl vous reste " + strconv.Itoa(characterInfo.skillPoints) + " point" + plural + " à répartir.\n"
	}

	return characterString
}

func monsterToString(monsterInfo monster) string {
	return monsterInfo.monsterName + strconv.Itoa(monsterInfo.currentHp) + " / " + strconv.Itoa(10+monsterInfo.currentHp+monsterInfo.constitution) + " HP\n"
}

func getMaxHP(characterInfo character) int {
	return 10 + characterInfo.constitution + characterInfo.level
}
