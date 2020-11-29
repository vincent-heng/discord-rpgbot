package main

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func fetchCharacters() (string, error) {
	// TODO update current hp, stamina and last update
	rows, err := db.Query("SELECT discord_id, level FROM character")
	if err != nil {
		return "", err
	}

	characters := ""
	for rows.Next() {
		var discordId int
		var level int
		err = rows.Scan(&discordId, &level)
		if err != nil {
			return "", err
		}
		characters += discordIdToText(discordId) + " (niv. " + strconv.Itoa(level) + ") "
	}
	return characters, nil
}

func fetchCharacterInfo(tx *sql.Tx, userId int) (character, error) {
	existingTransaction := tx != nil
	if !existingTransaction {
		var err error
		tx, err = db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		defer tx.Rollback()
	}
	stmt, err := tx.Prepare("SELECT discord_id, class, experience, level, strength, agility, wisdom, constitution, skill_points, current_hp, stamina FROM character WHERE discord_id = $1")
	if err != nil {
		return character{}, err
	}
	defer stmt.Close()

	characterInfo := character{}

	rows, err := stmt.Query(userId)
	found := false
	for rows.Next() {
		found = true
		var classString *string
		rows.Scan(&characterInfo.discordId, &classString, &characterInfo.experience, &characterInfo.level, &characterInfo.strength, &characterInfo.agility, &characterInfo.wisdom, &characterInfo.constitution, &characterInfo.skillPoints, &characterInfo.currentHp, &characterInfo.stamina)
		characterInfo.class = getClass(*classString)
	}

	if !existingTransaction {
		err = tx.Commit() // COMMIT TRANSACTION
		if err != nil {
			return character{}, err
		}
	}

	if !found {
		return character{}, nil
	}

	return characterInfo, nil
}

func fetchBattleParticipants(tx *sql.Tx, monsterQueueId int) ([]character, error) {
	existingTransaction := tx != nil
	if !existingTransaction {
		var err error
		tx, err = db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		defer tx.Rollback()
	}
	stmt, err := tx.Prepare("SELECT DISTINCT discord_id, experience, level, skill_points FROM character INNER JOIN battle_participation ON character.discord_id = battle_participation.character_discord_id WHERE monster_queue_id = $1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var participants []character

	rows, err := stmt.Query(monsterQueueId)
	for rows.Next() {
		characterInfo := character{}
		rows.Scan(&characterInfo.discordId, &characterInfo.experience, &characterInfo.level, &characterInfo.skillPoints)
		participants = append(participants, characterInfo)
	}

	if !existingTransaction {
		err = tx.Commit() // COMMIT TRANSACTION
		if err != nil {
			return nil, err
		}
	}

	return participants, nil
}

func fetchMonsterInfo(tx *sql.Tx) (monster, error) {
	existingTransaction := tx != nil
	if !existingTransaction {
		var err error
		tx, err = db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		defer tx.Rollback()
	}
	stmt, err := db.Prepare("SELECT monster_queue_id, monster_name, current_hp, agility, constitution, experience FROM monster_queue WHERE current_hp > 0 ORDER BY monster_queue_id LIMIT 1")
	if err != nil {
		return monster{}, err
	}
	defer stmt.Close()

	monsterInfo := monster{}
	rows, err := stmt.Query()

	found := false
	for rows.Next() {
		found = true
		rows.Scan(&monsterInfo.monsterId, &monsterInfo.monsterName, &monsterInfo.currentHp, &monsterInfo.agility, &monsterInfo.constitution, &monsterInfo.experience)
	}

	if !existingTransaction {
		err = tx.Commit() // COMMIT TRANSACTION
		if err != nil {
			return monster{}, err
		}
	}

	if !found {
		return monster{}, nil
	}

	return monsterInfo, nil
}

func createCharacter(discordId int) error {
	characterToCreate := getDefaultCharacter()
	characterToCreate.discordId = discordId

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO character(discord_id, class, experience, level, strength, agility, wisdom, constitution, skill_points, current_hp, stamina) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(characterToCreate.discordId, characterToCreate.class.String(), characterToCreate.experience, characterToCreate.level, characterToCreate.strength, characterToCreate.agility, characterToCreate.wisdom, characterToCreate.constitution, characterToCreate.skillPoints, characterToCreate.currentHp, characterToCreate.stamina)
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

func upStats(statsToUp string, userId int, amount int) error {

	tx, err := db.Begin() // BEGIN TRANSACTION
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	character, err := fetchCharacterInfo(tx, userId)
	if err != nil {
		return err // character unfetchable
	}

	if character.discordId == 0 {
		return errors.New("Character doesn't exist")
	}

	if amount > character.skillPoints {
		return errors.New("Not enough skill points")
	}

	switch statsToUp {
	case "strength":
		character.strength = character.strength + amount
	case "agility":
		character.agility = character.agility + amount
	case "wisdom":
		character.wisdom = character.wisdom + amount
	case "constitution":
		character.constitution = character.constitution + amount
	default:
		return errors.New("Wrong stat")
	}
	character.skillPoints = character.skillPoints - amount

	stmt, err := tx.Prepare("UPDATE character SET strength = $1, agility = $2, wisdom = $3, constitution = $4, skill_points = $5 WHERE discord_id = $6")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(character.strength, character.agility, character.wisdom, character.constitution, character.skillPoints, character.discordId)

	err = tx.Commit() // COMMIT TRANSACTION
	if err != nil {
		return err
	}

	return nil

}
