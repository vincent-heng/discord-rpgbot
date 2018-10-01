package main

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
)

func fetchCharacters() (string, error) {
	// TODO update current hp, stamina and last update
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

func fetchCharacterInfo(tx *sql.Tx, name string) (character, error) {
	existingTransaction := tx != nil
	if !existingTransaction {
		var err error
		tx, err = db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		defer tx.Rollback()
	}
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
	stmt, err := tx.Prepare("SELECT DISTINCT name, experience, level, skill_points FROM character INNER JOIN battle_participation ON character.name = battle_participation.character_name WHERE monster_queue_id = $1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var participants []character

	rows, err := stmt.Query(monsterQueueId)
	for rows.Next() {
		characterInfo := character{}
		rows.Scan(&characterInfo.name, &characterInfo.experience, &characterInfo.level, &characterInfo.skillPoints)
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

	tx, err := db.Begin() // BEGIN TRANSACTION
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	// Get Character
	attacker, err := fetchCharacterInfo(tx, characterName)
	if err != nil {
		return "", err
	}
	if attacker.name == "" {
		return "", errors.New("Character not found")
	}

	// TODO Check has enough stamina

	// Get Monster
	monsterTarget, err := fetchMonsterInfo(tx)
	if err != nil {
		return "", err
	}
	if monsterTarget.monsterName == "" {
		return "", errors.New("Monster not found")
	}

	// Compute fight
	agilityBonus := rand.Intn(attacker.agility*2 + 1)
	hitPoints := attacker.strength + agilityBonus
	damageReduction := monsterTarget.agility
	result := hitPoints - damageReduction
	if result <= 0 { // At least 1 damage
		result = 1
	}
	monsterTarget.currentHp = monsterTarget.currentHp - result

	// TODO Use stamina

	// Update monster's current hp
	stmt, err := tx.Prepare("UPDATE monster_queue SET current_hp = $1 where monster_queue_id = $2")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	_, err = stmt.Exec(monsterTarget.currentHp, monsterTarget.monsterId)

	// Add character to battle participation
	stmt, err = tx.Prepare("INSERT INTO battle_participation(monster_queue_id, character_name) VALUES($1, $2)")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	_, err = stmt.Exec(monsterTarget.monsterId, attacker.name)

	resultText := "**" + characterName + "** inflige " + strconv.Itoa(result) + " (" + strconv.Itoa(attacker.strength) + "+" + strconv.Itoa(agilityBonus) + "-" + strconv.Itoa(monsterTarget.agility) + ") points de dégâts à **" + monsterTarget.monsterName + "**.\n"

	if monsterTarget.currentHp <= 0 { // Target defeated
		resultText = resultText + "L'adversaire est vaincu ! Le combat rapporte " + strconv.Itoa(monsterTarget.experience) + " points d'expérience partagés entre :\n"
		// Gain XP for every participants
		participants, err := fetchBattleParticipants(tx, monsterTarget.monsterId)
		if err != nil {
			return "", err
		}

		sharedExperience := (monsterTarget.experience) / len(participants)

		for _, participant := range participants {
			resultText = resultText + "- " + participant.name
			participant.experience = participant.experience + sharedExperience
			newLevel := parseLevel(participant.experience)
			if participant.level < newLevel {
				nbLevelUps := newLevel - participant.level
				resultText = resultText + ": Gain de niveau ! "
				if nbLevelUps > 1 {
					resultText = resultText + " x" + strconv.Itoa(nbLevelUps)
				}
				participant.level = newLevel
				participant.skillPoints = participant.skillPoints + nbLevelUps*5
			}
			// update character: xp, level, skillPoints
			stmt, err := tx.Prepare("UPDATE character SET experience = $1, level = $2, skill_points = $3 WHERE name = $4")
			if err != nil {
				return "", err
			}
			defer stmt.Close()

			_, err = stmt.Exec(participant.experience, participant.level, participant.skillPoints, participant.name)

			resultText = resultText + "\n"
		}

	}

	err = tx.Commit() // COMMIT TRANSACTION
	if err != nil {
		return "", err
	}

	return resultText, nil
}

func upStats(statsToUp string, username string, amount int) error {

	tx, err := db.Begin() // BEGIN TRANSACTION
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	character, err := fetchCharacterInfo(tx, username)
	if err != nil {
		return err // character unfetchable
	}

	if character.name == "" {
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

	stmt, err := tx.Prepare("UPDATE character SET strength = $1, agility = $2, wisdom = $3, constitution = $4, skill_points = $5 WHERE name = $6")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(character.strength, character.agility, character.wisdom, character.constitution, character.skillPoints, character.name)

	err = tx.Commit() // COMMIT TRANSACTION
	if err != nil {
		return err
	}

	return nil

}
