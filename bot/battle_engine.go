package bot

import (
	"math"
	"math/rand"
	"strconv"

	"gorm.io/gorm"

	"github.com/vincent-heng/discord-airpgbot/bot/db"
	"github.com/vincent-heng/discord-airpgbot/bot/util"
)

func (b *Bot) attackCurrentMonster(characterID uint) (string, error) {
	tx := b.db.Begin()
	defer tx.Rollback()

	attacker, monster, err := tx.GetParticipants(characterID)
	if err != nil {
		return "", err
	}

	endOfFight := false
	actionReport := ""
	switch attacker.Class {
	case "Combattant":
		endOfFight, actionReport, err = b.triggerFighterAction(tx, attacker, monster)
		if err != nil {
			return "", err
		}
	}

	// Add character to battle participation
	if e := tx.Model(monster).Association("Participants").Append(attacker); e != nil {
		return "", e
	}

	if endOfFight { // Target defeated
		report, err := b.computeVictory(tx, monster)
		if err != nil {
			return "", err
		}

		actionReport += report
	}

	return actionReport, tx.Commit().Error
}

func parseLevel(experience int) int {
	floatExperience := float64(experience)
	floatLevel := (-1 + math.Sqrt(1+2*floatExperience/25)) / 2
	roundLevel := int(floatLevel) + 1 // 99 XP -> 0.9 floatLevel -> 1 roundLevel
	return roundLevel
}

func (b *Bot) computeVictory(tx *db.DB, monsterTarget *db.Monster) (string, error) {
	report := "L'adversaire est vaincu ! Le combat rapporte " +
		strconv.Itoa(monsterTarget.Experience) +
		" points d'expérience partagés entre :\n"

	// Gain XP for every participants
	var participants []db.Character
	tx.Model(monsterTarget).Association("Participants").Find(&participants)

	sharedExperience := (monsterTarget.Experience) / len(participants)

	for i := range participants {
		participant := participants[i]

		report += "- " + util.DiscordIDToText(participant.ID)
		participant.Experience = participant.Experience + sharedExperience
		newLevel := parseLevel(participant.Experience)
		if participant.Level < newLevel {
			nbLevelUps := newLevel - participant.Level
			report += ": Gain de niveau ! "
			if nbLevelUps > 1 {
				report += " x" + strconv.Itoa(nbLevelUps)
			}
			participant.Level = newLevel
			participant.SkillPoints = participant.SkillPoints + nbLevelUps*5
		}

		if err := tx.Save(&participant).Error; err != nil {
			return "", err
		}

		report += "\n"
	}

	return report, nil
}

func (b *Bot) triggerFighterAction(tx *db.DB, attacker *db.Character, monster *db.Monster) (bool, string, error) {
	agilityBonus := rand.Intn(attacker.Agility*2 + 1) //nolint:gosec
	hitPoints := attacker.Strength + agilityBonus
	damageReduction := monster.Agility
	damage := hitPoints - damageReduction
	if damage <= 0 { // At least 1 damage
		damage = 1
	}
	monster.CurrentHp = monster.CurrentHp - damage

	if e := tx.Model(&db.Monster{Model: gorm.Model{ID: monster.ID}}).Update("current_hp", monster.CurrentHp).Error; e != nil {
		return false, "", e
	}

	endOfFight := monster.CurrentHp <= 0
	actionReport := writeFighterActionReport(attacker, monster, damage, agilityBonus)
	return endOfFight, actionReport, nil
}

func writeFighterActionReport(attacker *db.Character, monster *db.Monster, damage int, agilityBonus int) string {
	return "**" +
		util.DiscordIDToText(attacker.ID) +
		"** inflige " +
		strconv.Itoa(damage) +
		" (" +
		strconv.Itoa(attacker.Strength) +
		"+" + strconv.Itoa(agilityBonus) +
		"-" +
		strconv.Itoa(monster.Agility) +
		") points de dégâts à **" +
		monster.Name +
		"**.\n"
}
