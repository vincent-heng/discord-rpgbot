package main

import (
	"errors"
	"log"
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
		characters += name + " (" + strconv.Itoa(experience) + ")"
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
		log.Fatal(err)
	}
	defer stmt.Close() // danger!
	rows, err := stmt.Query(name)
	rows.Next()
	var found int
	rows.Scan(&found)
	if found > 0 {
		err = tx.Commit()
		return errors.New("Character already exists")
	}

	stmt, err = tx.Prepare("INSERT character(name, experience) VALUES($1, 0)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close() // danger!
	stmt.Exec(name)

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
