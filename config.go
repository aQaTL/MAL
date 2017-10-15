package main

import (
	"os"
	"encoding/json"
	"log"
)

type Config struct {
	SelectedID        int
	MaxVisibleEntries int
}

func LoadConfig() (config *Config) {
	config = new(Config)
	config.MaxVisibleEntries = 10

	f, err := os.Open(ConfigFile)
	defer f.Close()
	if os.IsNotExist(err) {
		return
	} else if err != nil {
		log.Printf("Error opening %s file: %v", ConfigFile, err)
	}

	decoder := json.NewDecoder(f)
	decoder.Decode(config)

	return
}

func (cfg *Config) Save() {
	f, err := os.Create(ConfigFile)
	defer f.Close()
	if err != nil {
		log.Printf("Error saving %s file: %v", ConfigFile, err)
	}

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		log.Printf("Error encoding %s file: %v", ConfigFile, err)
	}
}
