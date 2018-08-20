package main

import (
	"encoding/json"
	"os"
)

// Configuration filled from configuration file
type Configuration struct {
	DiscordBotKey string
}

// loadConfiguration loads configuration from json file
func loadConfiguration(configurationFile string) (Configuration, error) {
	configuration := Configuration{}

	file, err := os.Open(configurationFile)
	if err != nil {
		return configuration, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	return configuration, err
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
