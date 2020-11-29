package main

import (
	"errors"
	"math/rand"
	"strconv"

	"gorm.io/gorm"
)

func attackCurrentMonster(characterDiscordID uint) (string, error) {
	var actionReport *string
	if e := db.Transaction(func(tx *gorm.DB) error {
		attacker, monster, err := getParticipants(tx, characterDiscordID)
		if err != nil {
			return err
		}

		endOfFight := false
		switch attacker.Class {
		case "Combattant":
			endOfFight, actionReport, err = triggerFighterAction(tx, attacker, monster)
			if err != nil {
				return err
			}
		}

		// Add character to battle participation
		err = updateBattleParticipation(tx, attacker, monster)
		if err != nil {
			return err
		}

		if endOfFight { // Target defeated
			err = computeVictory(tx, monster, actionReport)
			if err != nil {
				return err
			}
		}

		return nil
	}); e != nil {
		return "", e
	}

	return *actionReport, nil
}

func getParticipants(tx *gorm.DB, characterDiscordID uint) (*Character, *Monster, error) {
	attacker, err := fetchCharacterInfo(tx, characterDiscordID)
	if err != nil {
		return nil, nil, err
	}

	monsterTarget, err := fetchMonsterInfo(tx)
	if err == gorm.ErrRecordNotFound {
		return nil, nil, errors.New("Monster not found")
	}
	if err != nil {
		return nil, nil, err
	}

	return &attacker, &monsterTarget, nil
}

func updateBattleParticipation(tx *gorm.DB, attacker *Character, monsterTarget *Monster) error {
	monsterTarget.Participants = append(monsterTarget.Participants, attacker)
	if e := tx.Save(&monsterTarget).Error; e != nil {
		return e
	}
	return nil
}

func computeVictory(tx *gorm.DB, monsterTarget *Monster, actionReport *string) error {
	*actionReport = *actionReport + "L'adversaire est vaincu ! Le combat rapporte " + strconv.Itoa(monsterTarget.Experience) + " points d'expérience partagés entre :\n"
	// Gain XP for every participants
	sharedExperience := (monsterTarget.Experience) / len(monsterTarget.Participants)

	for _, participant := range monsterTarget.Participants {
		*actionReport = *actionReport + "- " + discordIDToText(participant.ID)
		participant.Experience = participant.Experience + sharedExperience
		newLevel := parseLevel(participant.Experience)
		if participant.Level < newLevel {
			nbLevelUps := newLevel - participant.Level
			*actionReport = *actionReport + ": Gain de niveau ! "
			if nbLevelUps > 1 {
				*actionReport = *actionReport + " x" + strconv.Itoa(nbLevelUps)
			}
			participant.Level = newLevel
			participant.SkillPoints = participant.SkillPoints + nbLevelUps*5
		}

		tx.Save(&participant)

		*actionReport = *actionReport + "\n"
	}
	return nil
}

func triggerFighterAction(tx *gorm.DB, attacker *Character, monsterTarget *Monster) (bool, *string, error) {
	agilityBonus := rand.Intn(attacker.Agility*2 + 1)
	hitPoints := attacker.Strength + agilityBonus
	damageReduction := monsterTarget.Agility
	damage := hitPoints - damageReduction
	if damage <= 0 { // At least 1 damage
		damage = 1
	}
	monsterTarget.CurrentHp = monsterTarget.CurrentHp - damage

	if e := tx.Model(&Monster{Model: gorm.Model{ID: monsterTarget.ID}}).Update("current_hp", monsterTarget.CurrentHp).Error; e != nil {
		return false, nil, e
	}

	endOfFight := monsterTarget.CurrentHp <= 0
	actionReport := writeFighterActionReport(attacker, monsterTarget, damage, agilityBonus)
	return endOfFight, &actionReport, nil
}

func writeFighterActionReport(attacker *Character, monsterTarget *Monster, damage int, agilityBonus int) string {
	return "**" + discordIDToText(attacker.ID) + "** inflige " + strconv.Itoa(damage) + " (" + strconv.Itoa(attacker.Strength) + "+" + strconv.Itoa(agilityBonus) + "-" + strconv.Itoa(monsterTarget.Agility) + ") points de dégâts à **" + monsterTarget.MonsterName + "**.\n"
}
