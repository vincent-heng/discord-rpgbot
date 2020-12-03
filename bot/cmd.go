package bot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"

	"github.com/vincent-heng/discord-airpgbot/bot/db"
	"github.com/vincent-heng/discord-airpgbot/bot/util"
)

func (b *Bot) charactersCmd(s *discordgo.Session, m *discordgo.MessageCreate, _ uint) _Response {
	characters, err := b.db.FetchCharacters()
	if err != nil {
		return simpleErr(err, "Impossible de récupérer la liste.")
	}
	return simpleResponse(characters)
}

func (b *Bot) joinAdventure(s *discordgo.Session, m *discordgo.MessageCreate, authorID uint) _Response {
	if err := b.db.CreateCharacter(authorID); err != nil {
		return simpleErr(fmt.Errorf("cannot create character: %w", err),
			"Impossible de créer le personnage...")
	}

	return simpleResponse(util.DiscordIDToText(authorID) + " a rejoint l'aventure !")
}

func (b *Bot) characterCmd(s *discordgo.Session, m *discordgo.MessageCreate, authorID uint) _Response {
	c, err := b.db.FetchCharacterInfo(authorID)
	if err != nil {
		return simpleErr(fmt.Errorf("cannot fetch character info: %w", err),
			"Impossible de récupérer les informations du personnage.")
	}

	if c.ID == 0 {
		return simpleErr(fmt.Errorf("id: %v, name: %v, err: %w", authorID, m.Author.Username, errCharacterDoesNotExist),
			"Vous devez d'abord rejoindre l'aventure en tapant !join_adventure")
	}

	return simpleResponse(c.String())
}

func (b *Bot) watchCmd(s *discordgo.Session, m *discordgo.MessageCreate, _ uint) _Response {
	monster, err := b.db.FetchMonsterInfo()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return simpleResponse("Il n'y a plus de monstre... pour l'instant !")
	}

	if err == nil {
		return simpleResponse(monster.String())
	}

	return simpleErr(err, "Impossible de récupérer les informations du monstre actuel.")
}

func (b *Bot) hitCmd(s *discordgo.Session, m *discordgo.MessageCreate, authorID uint) _Response {
	report, err := b.attackCurrentMonster(authorID)
	if err != nil {
		return simpleErr(fmt.Errorf("cannot attack monster: %w", err), "Impossible d'attaquer.")
	}

	return simpleResponse(report)
}

func (b *Bot) handleUpStats(s *discordgo.Session, m *discordgo.MessageCreate, userID uint, stat string) _Response {
	statTrigram := stat[0:3]
	content := strings.TrimSpace(strings.TrimPrefix(m.Content, "!"+statTrigram+" "))
	amount, err := strconv.Atoi(content)
	if err != nil {
		return simpleErr(fmt.Errorf("cannot upgrading stats: %w", errIllegalArgument),
			"Mauvaise syntaxe, essayez `!"+statTrigram+" 1`")
	}

	if amount <= 0 {
		return simpleErr(fmt.Errorf("cannot upgrading stats negative value: %w",
			errIllegalArgument), "Mauvaise syntaxe, essayez un nombre positif :unamused:")
	}

	if e := b.db.UpStats(stat, userID, amount); e != nil {
		return simpleErr(fmt.Errorf("cannot upgrade stat: %w", e), "Répartition impossible.")
	}

	return simpleResponse("Répartition effectuée !")
}

func (b *Bot) startAdventureCmd(s *discordgo.Session, m *discordgo.MessageCreate, _ uint) _Response {
	if err := util.SetAdventureChannel(m.ChannelID); err != nil {
		return simpleErr(fmt.Errorf("cannot set adventure: %w", err), "")
	}

	return simpleResponse("L'aventure commence ici.")
}

func (b *Bot) shoutCmd(s *discordgo.Session, m *discordgo.MessageCreate, _ uint) _Response {
	content := strings.TrimSpace(strings.TrimPrefix(m.Content, "!shout "))

	channelID, err := util.GetChannelID()
	if err != nil {
		return simpleErr(fmt.Errorf("cannot get channel ID: %w", err),
			"Error retrieving channel ID")
	}

	if channelID == "" {
		return simpleResponse("Set the channel with !start_adventure")
	}

	return _Response{
		msgs: []_Message{
			{Channel: channelID, Message: content},
		},
	}
}

func (b *Bot) spawnCmd(s *discordgo.Session, m *discordgo.MessageCreate, _ uint) _Response {
	content := strings.TrimSpace(strings.TrimPrefix(m.Content, "!spawn "))

	params := strings.Split(content, "_")
	if len(params) < 6 {
		return simpleErr(fmt.Errorf("syntax: Name of the mob_XP_str_agi_wis_con: %w", errIllegalArgument),
			"Bad arguments. Syntax: Name of the mob_XP_str_agi_wis_con")
	}

	_m := db.Monster{
		Name: params[0],
	}

	for i, ptr := range []*int{&_m.Experience, &_m.Strength, &_m.Agility, &_m.Wisdom, &_m.Constitution} {
		value, err := strconv.Atoi(params[i+1])
		if err != nil {
			// TODO improve msg
			return simpleErr(fmt.Errorf("cannot spawn monster: %w", err), "Illegal argument")

		}
		*ptr = value
	}

	if err := b.db.SpawnMonster(_m); err != nil {
		return simpleErr(fmt.Errorf("spawning monster: %w", err), "Error spawning monster")
	}

	return simpleResponse("Monster spawned")
}
