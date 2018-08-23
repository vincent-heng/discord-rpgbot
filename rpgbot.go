package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
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
			s.ChannelMessageSend(m.ChannelID, "Impossible de récupérer la liste.")
		} else {
			log.Printf("[Response] %v", characters)
			s.ChannelMessageSend(m.ChannelID, characters)
		}
	} else if m.Content == "!join_adventure" {
		log.Printf("[Request] Join adventure: %v", m.Author.Username)
		err := createCharacter(m.Author.Username)
		if err != nil {
			log.Printf("[Response] %v", err)
			s.ChannelMessageSend(m.ChannelID, "Impossible de créer le personnage...")
		} else {
			log.Printf("[Response] %v joined the adventure!", m.Author.Username)
			s.ChannelMessageSend(m.ChannelID, m.Author.Username+" a rejoint l'aventure !")
		}
	} else if m.Content == "!character" {
		log.Printf("[Request] Character info: %v", m.Author.Username)
		characterInfo, err := fetchCharacterInfo(m.Author.Username)
		if err != nil {
			log.Printf("[Response] Error fetching character info: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Impossible de récupérer les informations du personnage.")
			return
		}
		if characterInfo == "" {
			log.Printf("[Response] Character doesn't exist")
			s.ChannelMessageSend(m.ChannelID, "Vous devez d'abord rejoindre l'aventure en tapant !join_adventure")
			return
		}
		log.Printf("[Response] %v", characterInfo)
		s.ChannelMessageSend(m.ChannelID, characterInfo)
	} else if m.Content == "!watch" {
		log.Println("[Request] Current monster info")
		monsterInfo, err := fetchMonsterInfo()
		if err != nil {
			log.Printf("[Response] Error fetching monster info: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Impossible de récupérer les informations du monstre actuel.")
			return
		}
		if monsterInfo == "" {
			log.Printf("[Response] Il n'y a plus de monstre... pour l'instant !")
			s.ChannelMessageSend(m.ChannelID, "Vous devez d'abord rejoindre l'aventure en tapant !join_adventure")
			return
		}

		log.Printf("[Response] %v", monsterInfo)
		s.ChannelMessageSend(m.ChannelID, monsterInfo)

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
	} else if strings.HasPrefix(m.Content, "!spawn ") {
		content := strings.TrimSpace(strings.TrimPrefix(m.Content, "!spawn "))
		log.Printf("[Request GM] Spawn: %v", content)
		params := strings.Split(content, "_")
		if len(params) < 3 {
			log.Println("[Response GM] Bad arguments. Syntax: Name of the mob_HP_XP")
			s.ChannelMessageSend(m.ChannelID, "Bad arguments. Syntax: Name of the mob_HP_XP")
			return
		}
		healthPoints, err := strconv.Atoi(params[1])
		if err != nil {
			log.Printf("[Response GM] HP should be an integer")
			s.ChannelMessageSend(m.ChannelID, "HP should be an integer")
		}
		experience, err := strconv.Atoi(params[2])
		if err != nil {
			log.Printf("[Response GM] HP should be an integer")
			s.ChannelMessageSend(m.ChannelID, "HP should be an integer")
		}
		err = spawnMonster(params[0], healthPoints, experience)
		if err != nil {
			log.Printf("[Response GM] Error spawning monster: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Error spawning monster")
		}
		s.ChannelMessageSend(m.ChannelID, "Monster spawned")
	}
}
