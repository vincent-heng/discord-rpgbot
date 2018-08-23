package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func fetchCharacters() (string, error) {
	rows, err := db.Query("SELECT name, level FROM character")
	if err != nil {
		return "", err
	}

	characters := ""
	for rows.Next() {
		var name string
		var level int
		err = rows.Scan(&name, &level)
		if err != nil {
			return "", err
		}
		characters += name + " (niv. " + strconv.Itoa(level) + ") "
	}
	return characters, nil
}

func fetchCharacterInfo(name string) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("SELECT name, class, experience, level, strength, agility, wisdom, constitution, skill_points, current_hp FROM character WHERE name ~* $1")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	characterInfo := character{}

	rows, err := stmt.Query(name)
	found := false
	for rows.Next() {
		found = true
		rows.Scan(&characterInfo.name, &characterInfo.class, &characterInfo.experience, &characterInfo.level, &characterInfo.strength, &characterInfo.agility, &characterInfo.wisdom, &characterInfo.constitution, &characterInfo.skillPoints, &characterInfo.current_hp)
	}

	if !found {
		return "", nil
	}

	return characterToString(characterInfo), nil
}

func characterToString(characterInfo character) string {
	characterString := characterInfo.name + " (" + characterInfo.class + ") - " + strconv.Itoa(characterInfo.current_hp) + " / " + strconv.Itoa(getMaxHP(characterInfo)) + " HP\n"
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

func getMaxHP(characterInfo character) int {
	return 10 + characterInfo.constitution + characterInfo.level
}

func createCharacter(name string) error {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("SELECT COUNT(name) FROM character where name ~* $1")
	if err != nil {
		return err
	}
	defer stmt.Close()
	rows, err := stmt.Query(name)
	for rows.Next() {
		var found int
		rows.Scan(&found)
		if found > 0 {
			err = tx.Commit()
			return errors.New("Character already exists")
		}
	}

	stmt, err = tx.Prepare("INSERT INTO character(name, class, experience, level, strength, agility, wisdom, constitution, skill_points, current_hp) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, "Combattant", 0, 1, 1, 1, 1, 1, 5, 12)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
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

func spawnMonster(monsterName string, healthPoints int, experience int) error {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("INSERT INTO monster_queue(monster_name, current_hp, max_hp, experience) VALUES ($1, $2, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(monsterName, healthPoints, experience)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
