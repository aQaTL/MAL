package main

import (
	"os"
	"encoding/json"
	"log"
	"github.com/aqatl/mal/mal"
	"time"
	"fmt"
)

type Config struct {
	SelectedID           int
	MaxVisibleEntries    int
	Websites             map[int]string
	Status               mal.MyStatus
	StatusAutoUpdateMode StatusAutoUpdateMode
	Sorting              Sorting
	LastUpdate           time.Time
}

type StatusAutoUpdateMode byte

const (
	Off            StatusAutoUpdateMode = iota
	Normal
	AfterThreshold
)

type Sorting byte

const (
	ByLastUpdated     Sorting = iota
	ByTitle
	ByWatchedEpisodes
	ByScore
)

func ParseSorting(sort string) (Sorting, error) {
	var sorting Sorting

	switch sort {
	case "last-updated":
		sorting = ByLastUpdated
	case "title":
		sorting = ByTitle
	case "episodes":
		sorting = ByWatchedEpisodes
	case "score":
		sorting = ByScore
	default:
		return 0, fmt.Errorf("invalid option; possible values: " +
			"last-updated|title|episodes|score")
	}

	return sorting, nil
}

func LoadConfig() (c *Config) {
	c = new(Config)
	c.MaxVisibleEntries = 10
	c.Websites = make(map[int]string)
	c.Status = mal.All
	c.StatusAutoUpdateMode = Off
	c.Sorting = ByLastUpdated

	f, err := os.Open(ConfigFile)
	defer f.Close()
	if os.IsNotExist(err) {
		return
	} else if err != nil {
		log.Printf("Error opening %s file: %v", ConfigFile, err)
	}

	decoder := json.NewDecoder(f)
	decoder.Decode(c)

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
