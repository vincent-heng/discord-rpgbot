package bot

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/vincent-heng/discord-airpgbot/bot/db"
	"github.com/vincent-heng/discord-airpgbot/bot/util"
	"github.com/vincent-heng/discord-airpgbot/config"
)

// Bot is the discord bot manager
type Bot struct {
	config.Config

	db *db.DB
}

type _Message struct {
	Channel string
	Message string
}

func (m _Message) getChan(id string) string {
	if m.Channel == "" {
		return id
	}
	return m.Channel
}

type _Response struct {
	msgs []_Message
	err  error
}

func simpleErr(err error, msg string) _Response {
	if msg == "" && err != nil {
		msg = err.Error()
	}

	return _Response{
		err: err,
		msgs: []_Message{
			{Message: msg},
		},
	}
}

func simpleResponse(msg string) _Response {
	return _Response{
		msgs: []_Message{
			{Message: msg},
		},
	}
}

type _Handler func(*Bot, *discordgo.Session, *discordgo.MessageCreate, uint) _Response

// New instantiates a bot with config
func New(conf config.Config) (*Bot, error) {
	database, err := db.New()
	if err != nil {
		return nil, err
	}

	return &Bot{
		Config: conf,
		db:     database,
	}, nil
}

var (
	// cmd router
	router = map[string]_Handler{ //nolint:gochecknoglobals
		"characters":     (*Bot).charactersCmd,
		"join_adventure": (*Bot).joinAdventure,
		"character":      (*Bot).characterCmd,
		"watch":          (*Bot).watchCmd,
		"hit":            (*Bot).hitCmd,
		"str":            handleUpStatsFunctor("strength"),
		"agi":            handleUpStatsFunctor("agility"),
		"wis":            handleUpStatsFunctor("wisdom"),
		"con":            handleUpStatsFunctor("constitution"),
		// game master cmd
		"start_adventure": gameMasterCmdFunctor((*Bot).startAdventureCmd),
		"shout":           gameMasterCmdFunctor((*Bot).shoutCmd),
		"spawn":           gameMasterCmdFunctor((*Bot).spawnCmd),
	}
)

func gameMasterCmdFunctor(handler _Handler) _Handler {
	return func(b *Bot, s *discordgo.Session, m *discordgo.MessageCreate, authorID uint) _Response {
		// GM commands
		if authorID != b.Config.GameMaster {
			return simpleErr(errNotGameMaster, "")
		}
		return handler(b, s, m, authorID)
	}
}

func handleUpStatsFunctor(stat string) _Handler {
	return func(b *Bot, s *discordgo.Session, m *discordgo.MessageCreate, userID uint) _Response {
		return b.handleUpStats(s, m, userID, stat)
	}
}

// Handler for discord events
func (b *Bot) Handler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	content := strings.Split(m.Content, " ")
	if len(content) < 1 {
		return
	}

	authorID64, err := strconv.ParseUint(strings.TrimSpace(m.Author.ID), 10, 64)
	if err != nil {
		log.Warn().Msg("[Response] Unexpected error (authorID not an integer)")
		s.ChannelMessageSend(m.ChannelID, "Erreur inattendue :cry:")
	}
	authorID := uint(authorID64)

	channelID, err := util.GetChannelID()
	if err != nil {
		log.Error().Err(err).Msg("[Response]")
		return
	}

	handler, ok := router[content[0][1:]]
	if !ok {
		// not a cmd
		return
	}

	if channelID != m.ChannelID && authorID != b.Config.GameMaster {
		log.Warn().Str("expected", channelID).Str("current", m.ChannelID).Msg("request on a wrong channel")
		// return
	}

	uuid := uuid.New().String()
	log.Debug().
		Str("cmd", content[0]).
		Uint("user", authorID).
		Strs("params", content[1:]).
		Str("cmdID", uuid).
		Msg("calling handler for cmd")

	resp := handler(b, s, m, authorID)

	for i := range resp.msgs {
		msg := &resp.msgs[i]

		if _, err := s.ChannelMessageSend(msg.getChan(m.ChannelID), msg.Message); err != nil {
			log.Error().Err(err).Msg("cannot push message")
		}
	}

	log.Debug().
		Str("cmdID", uuid).
		Err(resp.err).
		Interface("message", resp.msgs).
		Msg("cmd done")
}
