package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func fetchCharacters() (string, error) {
	rows, err := db.Query("SELECT name, experience FROM character")
	if err != nil {
		return "", err
	}

	characters := ""
	for rows.Next() {
		var name string
		var experience int
		err = rows.Scan(&name, &experience)
		if err != nil {
			return "", err
		}
		characters += name + " (" + strconv.Itoa(experience) + ") "
	}
	return characters, nil
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

	stmt, err = tx.Prepare("INSERT INTO character(name, experience) VALUES($1, 0)")
	if err != nil {
		return err
	}
	defer stmt.Close() // danger!
	_, err = stmt.Exec(name)
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
