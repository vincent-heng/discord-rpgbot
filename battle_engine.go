package main

import (
	"database/sql"
	"errors"
	"log"
	"math/rand"
	"strconv"
)

func attackCurrentMonster(characterDiscordId int) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	attacker, monster, err := getParticipants(tx, characterDiscordId)
	if err != nil {
		return "", err
	}

	endOfFight := false
	var actionReport *string
	switch attacker.class {
	case FIGHTER:
		endOfFight, actionReport, err = triggerFighterAction(tx, attacker, monster)
		if err != nil {
			return "", err
		}
	}

	// Add character to battle participation
	err = updateBattleParticipation(tx, attacker, monster)
	if err != nil {
		return "", err
	}

	if endOfFight { // Target defeated
		err = computeVictory(tx, monster, actionReport)
		if err != nil {
			return "", err
		}
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return *actionReport, nil
}

func getParticipants(tx *sql.Tx, characterDiscordId int) (*character, *monster, error) {
	attacker, err := fetchCharacterInfo(tx, characterDiscordId)
	if err != nil {
		return nil, nil, err
	}
	if attacker.discordId == 0 {
		return nil, nil, errors.New("Character not found")
	}

	monsterTarget, err := fetchMonsterInfo(tx)
	if err != nil {
		return nil, nil, err
	}
	if monsterTarget.monsterName == "" {
		return nil, nil, errors.New("Monster not found")
	}
	return &attacker, &monsterTarget, nil
}

func updateBattleParticipation(tx *sql.Tx, attacker *character, monsterTarget *monster) error {
	stmt, err := tx.Prepare("INSERT INTO battle_participation(monster_queue_id, character_discord_id) VALUES($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(monsterTarget.monsterId, attacker.discordId)
	if err != nil {
		return err
	}
	return nil
}

func computeVictory(tx *sql.Tx, monsterTarget *monster, actionReport *string) error {
	*actionReport = *actionReport + "L'adversaire est vaincu ! Le combat rapporte " + strconv.Itoa(monsterTarget.experience) + " points d'expérience partagés entre :\n"
	// Gain XP for every participants
	participants, err := fetchBattleParticipants(tx, monsterTarget.monsterId)
	if err != nil {
		return err
	}

	sharedExperience := (monsterTarget.experience) / len(participants)

	for _, participant := range participants {
		*actionReport = *actionReport + "- " + discordIdToText(participant.discordId)
		participant.experience = participant.experience + sharedExperience
		newLevel := parseLevel(participant.experience)
		if participant.level < newLevel {
			nbLevelUps := newLevel - participant.level
			*actionReport = *actionReport + ": Gain de niveau ! "
			if nbLevelUps > 1 {
				*actionReport = *actionReport + " x" + strconv.Itoa(nbLevelUps)
			}
			participant.level = newLevel
			participant.skillPoints = participant.skillPoints + nbLevelUps*5
		}
		// update character: xp, level, skillPoints
		stmt, err := tx.Prepare("UPDATE character SET experience = $1, level = $2, skill_points = $3 WHERE discord_id = $4")
		if err != nil {
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(participant.experience, participant.level, participant.skillPoints, participant.discordId)

		*actionReport = *actionReport + "\n"
	}
	return nil
}

func triggerFighterAction(tx *sql.Tx, attacker *character, monsterTarget *monster) (bool, *string, error) {
	agilityBonus := rand.Intn(attacker.agility*2 + 1)
	hitPoints := attacker.strength + agilityBonus
	damageReduction := monsterTarget.agility
	damage := hitPoints - damageReduction
	if damage <= 0 { // At least 1 damage
		damage = 1
	}
	monsterTarget.currentHp = monsterTarget.currentHp - damage

	// Update monster's current hp
	stmt, err := tx.Prepare("UPDATE monster_queue SET current_hp = $1 where monster_queue_id = $2")
	if err != nil {
		return false, nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(monsterTarget.currentHp, monsterTarget.monsterId)
	if err != nil {
		return false, nil, err
	}

	endOfFight := monsterTarget.currentHp <= 0
	actionReport := writeFighterActionReport(attacker, monsterTarget, damage, agilityBonus)
	return endOfFight, &actionReport, nil
}

func writeFighterActionReport(attacker *character, monsterTarget *monster, damage int, agilityBonus int) string {
	return "**" + discordIdToText(attacker.discordId) + "** inflige " + strconv.Itoa(damage) + " (" + strconv.Itoa(attacker.strength) + "+" + strconv.Itoa(agilityBonus) + "-" + strconv.Itoa(monsterTarget.agility) + ") points de dégâts à **" + monsterTarget.monsterName + "**.\n"
}
