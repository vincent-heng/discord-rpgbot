package main

import (
	"fmt"
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
	log.Printf("Starting")

	// Configuration
	conf, err := loadConfiguration("config.json")
	if err != nil {
		log.Fatal("Can't load config file:", err)
	}
	configuration = &conf

	// Database
	db, err = initDb()
	checkErr(err)
	defer db.Close()

	err = db.Ping()
	checkErr(err)

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
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func initDb() (*sql.DB, error) {
	DB_HOST := os.Getenv("DB_HOST")
	DB_PORT := os.Getenv("DB_PORT")
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")

	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
	log.Printf("%v", dbinfo)
	return sql.Open("postgres", dbinfo)
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "!echo ") {
		name := strings.TrimSpace(strings.TrimPrefix(m.Content, "!echo "))
		log.Printf("Request: %v", name)

		log.Printf("Response: %v", name)

		// Send
		s.ChannelMessageSend(m.ChannelID, name)
	} else if m.Content == "!characters" {
		fmt.Println("# Querying")
		rows, err := db.Query("SELECT * FROM character")
		checkErr(err)

		for rows.Next() {
			var name string
			var experience int
			err = rows.Scan(&name, &experience)
			checkErr(err)
			fmt.Printf("%v | %v \n", name, experience)
		}
	}
}
