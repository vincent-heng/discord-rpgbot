package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	_ "gorm.io/driver/postgres"

	"github.com/vincent-heng/discord-airpgbot/bot"
	"github.com/vincent-heng/discord-airpgbot/config"
)

// loadConfiguration loads configuration from json file
func loadConfiguration(configurationFile string) (config.Config, error) {
	configuration := config.Config{}

	file, err := os.Open(configurationFile)
	if err != nil {
		return configuration, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	return configuration, err
}

func main() {
	log.Info().Msg("Starting...")

	// Configuration
	conf, err := loadConfiguration("config.json")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config file")
	}

	// Discord
	dg, err := discordgo.New("Bot " + conf.DiscordBotKey)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create Discord session")
	}

	bot, err := bot.New(conf)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot init the bot")
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(bot.Handler)

	// Open a websocket connection to Discord and begin listening.
	if err := dg.Open(); err != nil {
		log.Fatal().Err(err).Msg("cannot open discord connection")
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Info().Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	if err := dg.Close(); err != nil {
		log.Error().Err(err).Msg("cannot close discrod connection")
	}
}
