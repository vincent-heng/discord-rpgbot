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

	content := strings.Split(m.Content, " ")
	if len(content) < 1 {
		return
	}

	switch content[0] {
	case "!characters":
		log.Println("[Request] List characters")
		characters, err := fetchCharacters()
		if err != nil {
			log.Printf("[Response] DB is unavailable")
			s.ChannelMessageSend(m.ChannelID, "Impossible de récupérer la liste.")
		} else {
			log.Printf("[Response] %v", characters)
			s.ChannelMessageSend(m.ChannelID, characters)
		}
	case "!join_adventure":
		log.Printf("[Request] Join adventure: %v", m.Author.Username)
		err := createCharacter(m.Author.Username)
		if err != nil {
			log.Printf("[Response] %v", err)
			s.ChannelMessageSend(m.ChannelID, "Impossible de créer le personnage...")
		} else {
			log.Printf("[Response] %v joined the adventure!", m.Author.Username)
			s.ChannelMessageSend(m.ChannelID, m.Author.Username+" a rejoint l'aventure !")
		}
	case "!character":
		log.Printf("[Request] Character info: %v", m.Author.Username)
		characterInfo, err := fetchCharacterInfo(nil, m.Author.Username)
		if err != nil {
			log.Printf("[Response] Error fetching character info: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Impossible de récupérer les informations du personnage.")
			return
		}
		if characterInfo.name == "" {
			log.Printf("[Response] Character doesn't exist")
			s.ChannelMessageSend(m.ChannelID, "Vous devez d'abord rejoindre l'aventure en tapant !join_adventure")
			return
		}
		characterInfoString := characterToString(characterInfo)
		log.Printf("[Response] %v", characterInfoString)
		s.ChannelMessageSend(m.ChannelID, characterInfoString)
	case "!watch":
		log.Println("[Request] Current monster info")
		monsterInfo, err := fetchMonsterInfo(nil)
		if err != nil {
			log.Printf("[Response] Error fetching monster info: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Impossible de récupérer les informations du monstre actuel.")
			return
		}
		if monsterInfo.monsterName == "" {
			log.Printf("[Response] No monster left")
			s.ChannelMessageSend(m.ChannelID, "Il n'y a plus de monstre... pour l'instant !")
			return
		}
		monsterInfoString := monsterToString(monsterInfo)

		log.Printf("[Response] %v", monsterInfoString)
		s.ChannelMessageSend(m.ChannelID, monsterInfoString)
	case "!hit":
		log.Printf("[Request] Attack from %v", m.Author.Username)
		report, err := attackCurrentMonster(m.Author.Username)
		if err != nil { // Character exists? Monster exists? Enough Stamina?
			log.Printf("[Response] Error Attacking monster: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Impossible d'attaquer.")
			return
		}
		log.Printf("[Response] %v Attacked", m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, report)
	case "!str":
		handleUpStats(s, m.ChannelID, "strength", m.Author.Username, m.Content)
	case "!agi":
		handleUpStats(s, m.ChannelID, "agility", m.Author.Username, m.Content)
	case "!wis":
		handleUpStats(s, m.ChannelID, "wisdom", m.Author.Username, m.Content)
	case "!con":
		handleUpStats(s, m.ChannelID, "constitution", m.Author.Username, m.Content)
	}

	// GM commands
	if m.Author.Username != configuration.GameMaster {
		return
	}

	switch content[0] {
	case "!start_adventure":
		log.Printf("[Request GM] Set adventure on channel: %v", m.ChannelID)
		err := setAdventureChannel(m.ChannelID)
		if err != nil {
			log.Printf("[Response GM] Can't set adventure: %v", err)
			return
		}
		log.Printf("[Response GM] Adventure set on channel: %v", m.ChannelID)
		s.ChannelMessageSend(m.ChannelID, "L'aventure commence ici.")
	case "!shout":
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
	case "!spawn":
		content := strings.TrimSpace(strings.TrimPrefix(m.Content, "!spawn "))
		log.Printf("[Request GM] Spawn: %v", content)
		params := strings.Split(content, "_")
		if len(params) < 6 {
			log.Println("[Response GM] Bad arguments. Syntax: Name of the mob_XP_str_agi_wis_con")
			s.ChannelMessageSend(m.ChannelID, "Bad arguments. Syntax: Name of the mob_XP_str_agi_wis_con")
			return
		}
		experience, err := strconv.Atoi(params[1])
		if err != nil {
			log.Printf("[Response GM] XP should be an integer")
			s.ChannelMessageSend(m.ChannelID, "XP should be an integer")
		}
		strength, err := strconv.Atoi(params[2])
		if err != nil {
			log.Printf("[Response GM] STR should be an integer")
			s.ChannelMessageSend(m.ChannelID, "STR should be an integer")
		}
		agility, err := strconv.Atoi(params[3])
		if err != nil {
			log.Printf("[Response GM] AGI should be an integer")
			s.ChannelMessageSend(m.ChannelID, "AGI should be an integer")
		}
		wisdom, err := strconv.Atoi(params[4])
		if err != nil {
			log.Printf("[Response GM] WIS should be an integer")
			s.ChannelMessageSend(m.ChannelID, "WIS should be an integer")
		}
		constitution, err := strconv.Atoi(params[5])
		if err != nil {
			log.Printf("[Response GM] CON should be an integer")
			s.ChannelMessageSend(m.ChannelID, "CON should be an integer")
		}

		monsterToSpawn := monster{}
		monsterToSpawn.monsterName = params[0]
		monsterToSpawn.experience = experience
		monsterToSpawn.strength = strength
		monsterToSpawn.agility = agility
		monsterToSpawn.wisdom = wisdom
		monsterToSpawn.constitution = constitution
		monsterToSpawn.currentHp = getMaxHPMonster(monsterToSpawn)

		err = spawnMonster(monsterToSpawn)
		if err != nil {
			log.Printf("[Response GM] Error spawning monster: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Error spawning monster")
		}
		s.ChannelMessageSend(m.ChannelID, "Monster spawned")
	}
}

func handleUpStats(s *discordgo.Session, channelID string, stat string, username string, message string) {
	statTrigram := stat[0:3]
	log.Printf("[Request] Up %v stats for %v", stat, username)
	content := strings.TrimSpace(strings.TrimPrefix(message, "!"+statTrigram+" "))
	amount, err := strconv.Atoi(content)
	if err != nil {
		log.Printf("[Response] Illegal argument for upgrading stats: %v", err)
		s.ChannelMessageSend(channelID, "Mauvaise syntaxe, essayez `!"+statTrigram+" 1`")
		return
	}

	if amount <= 0 {
		log.Printf("[Response] Illegal argument for upgrading stats: negative value")
		s.ChannelMessageSend(channelID, "Mauvaise syntaxe, essayez un nombre positif :unamused:")
		return
	}

	err = upStats(stat, username, amount)
	if err != nil {
		log.Printf("[Response] Error upgrading stat: %v", err)
		s.ChannelMessageSend(channelID, "Répartition impossible.")
		return
	}

	log.Printf("[Response] %v successfully upgraded %v stats", stat, username)
	s.ChannelMessageSend(channelID, "Répartition effectuée !")
}
