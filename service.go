package main

import (
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
)

func fetchCharacters() (string, error) {
	// TODO update current hpn stamina and last update
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

func fetchCharacterInfo(name string) (character, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("SELECT name, class, experience, level, strength, agility, wisdom, constitution, skill_points, current_hp, stamina FROM character WHERE name ~* $1")
	if err != nil {
		return character{}, err
	}
	defer stmt.Close()

	characterInfo := character{}

	rows, err := stmt.Query(name)
	found := false
	for rows.Next() {
		found = true
		rows.Scan(&characterInfo.name, &characterInfo.class, &characterInfo.experience, &characterInfo.level, &characterInfo.strength, &characterInfo.agility, &characterInfo.wisdom, &characterInfo.constitution, &characterInfo.skillPoints, &characterInfo.currentHp, &characterInfo.stamina)
	}

	if !found {
		return character{}, nil
	}

	return characterInfo, nil
}

func fetchMonsterInfo() (monster, error) {
	rows, err := db.Query("SELECT monster_queue_id, monster_name, current_hp, agility, constitution FROM monster_queue WHERE current_hp > 0 ORDER BY monster_queue_id LIMIT 1")
	if err != nil {
		return monster{}, err
	}

	monsterInfo := monster{}

	found := false
	for rows.Next() {
		found = true
		rows.Scan(&monsterInfo.monsterId, &monsterInfo.monsterName, &monsterInfo.currentHp, &monsterInfo.agility, &monsterInfo.constitution)
	}

	if !found {
		return monster{}, nil
	}

	return monsterInfo, nil
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

	characterToCreate := getDefaultCharacter()
	characterToCreate.name = name

	stmt, err = tx.Prepare("INSERT INTO character(name, class, experience, level, strength, agility, wisdom, constitution, skill_points, current_hp, stamina) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(characterToCreate.name, characterToCreate.class, characterToCreate.experience, characterToCreate.level, characterToCreate.strength, characterToCreate.agility, characterToCreate.wisdom, characterToCreate.constitution, characterToCreate.skillPoints, characterToCreate.currentHp, characterToCreate.stamina)
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

func spawnMonster(monsterToSpawn monster) error {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO monster_queue(monster_name, experience, strength, agility, wisdom, constitution, current_hp) VALUES ($1, $2, $3, $4, $5, $6, $7)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(monsterToSpawn.monsterName, monsterToSpawn.experience, monsterToSpawn.strength, monsterToSpawn.agility, monsterToSpawn.wisdom, monsterToSpawn.constitution, monsterToSpawn.currentHp)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func attackCurrentMonster(characterName string) (string, error) {
	//Monster exists? Enough Stamina?
	// Get Character
	attacker, err := fetchCharacterInfo(characterName)
	if err != nil {
		return "", err
	}
	if attacker.name == "" {
		return "", errors.New("Character not found")
	}

	// TODO Check has enough stamina

	// Get Monster
	monsterTarget, err := fetchMonsterInfo()
	if err != nil {
		return "", err
	}
	if monsterTarget.monsterName == "" {
		return "", errors.New("Monster not found")
	}

	agilityBonus := rand.Intn(attacker.agility + 1)
	hitPoints := attacker.strength + agilityBonus
	damageReduction := monsterTarget.agility
	result := hitPoints - damageReduction
	if result <= 0 { // At least 1 damage
		result = 1
	}
	monsterTarget.currentHp = monsterTarget.currentHp - result

	// TODO Use stamina
	// TODO Check if monsterTarget is dead
	// TODO 	Gain XP and Check if level up

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	// Update monster's current hp
	stmt, err := tx.Prepare("UPDATE monster_queue SET current_hp = $1 where monster_queue_id = $2")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	_, err = stmt.Exec(monsterTarget.currentHp, monsterTarget.monsterId)

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return characterName + " inflige " + strconv.Itoa(result) + " (" + strconv.Itoa(attacker.strength) + "+" + strconv.Itoa(agilityBonus) + "-" + strconv.Itoa(monsterTarget.agility) + ") points de dégâts au monstre.", nil

}
