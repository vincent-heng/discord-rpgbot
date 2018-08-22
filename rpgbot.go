package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"database/sql"

	"github.com/bwmarrin/discordgo"
	_ "github.com/lib/pq"
)

var (
	configuration *Configuration
	db            *sql.DB
)

const (
	DB_NAME = "rpg"
)

func main() {
	log.Printf("Starting...")

	// Configuration
	conf, err := loadConfiguration("config.json")
	if err != nil {
		log.Fatal("Can't load config file:", err)
	}
	configuration = &conf

	// Database
	db, err = initDb()
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer db.Close()

	// Discord
	dg, err := discordgo.New("Bot " + configuration.DiscordBotKey)
	if err != nil {
		log.Fatalf("error creating Discord session: %v", err)
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!characters" {
		log.Println("[Request] List characters")
		characters, err := fetchCharacters()
		if err != nil {
			log.Printf("[Response] DB is unavailable")
			s.ChannelMessageSend(m.ChannelID, "DB is unavailable")
		} else {
			log.Printf("[Response] %v", characters)
			s.ChannelMessageSend(m.ChannelID, characters)
		}
	} else if m.Content == "!join_adventure" {
		log.Printf("[Request] Join adventure: %v", m.Author.Username)
		err := createCharacter(m.Author.Username)
		if err != nil {
			log.Printf("[Response] %v", err)
			s.ChannelMessageSend(m.ChannelID, "Can't create character")
		} else {
			log.Printf("[Response] %v joined the adventure!", m.Author.Username)
			s.ChannelMessageSend(m.ChannelID, m.Author.Username+" joined the adventure!")
		}
	}

	// GM commands
	if m.Author.Username != configuration.GameMaster {
		return
	}

	if m.Content == "!start_adventure" {
		log.Printf("[Request GM] Set adventure on channel: %v", m.ChannelID)
		err := setAdventureChannel(m.ChannelID)
		if err != nil {
			log.Printf("[Response GM] Can't set adventure: %v", err)
			return
		}
		log.Printf("[Response GM] Adventure set on channel: %v", m.ChannelID)
		s.ChannelMessageSend(m.ChannelID, "The adventure starts here")
	}

	if strings.HasPrefix(m.Content, "!shout ") {
		content := strings.TrimSpace(strings.TrimPrefix(m.Content, "!shout "))
		log.Printf("[Request GM] Shout: %v", content)

		channelId, err := getChannelId()
		if err != nil {
			log.Printf("Error retrieving channel ID: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Error retrieving channel ID")
			return
		}
		if channelId == "" {
			s.ChannelMessageSend(m.ChannelID, "Set the channel with !start_adventure")
			return
		}
		log.Printf("[Response GM] %v", content)
		s.ChannelMessageSend(channelId, content)
	}
}
