package main

import (
	"encoding/json"
	"fmt"
	"github.com/aqatl/mal/anilist"
	"github.com/aqatl/mal/mal"
	"log"
	"os"
	"time"
)

type Config struct {
	Websites             map[int]string

	MaxVisibleEntries    int
	StatusAutoUpdateMode StatusAutoUpdateMode
	Sorting              Sorting
	LastUpdate           time.Time

	BrowserPath          string
	TorrentClientPath    string
	TorrentClientArgs    string

	SelectedID int
	Status     mal.MyStatus

	ALSelectedID int
	ALStatus     anilist.MediaListStatus
}

type StatusAutoUpdateMode byte

const (
	Off StatusAutoUpdateMode = iota
	Normal
	AfterThreshold
)

type Sorting byte

const (
	ByLastUpdated Sorting = iota
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
	c = &Config{
		Websites:             make(map[int]string),

		MaxVisibleEntries:    20,
		StatusAutoUpdateMode: Off,
		Sorting:              ByLastUpdated,

		TorrentClientPath:    "qbittorrent",

		Status:               mal.All,

		ALStatus:             anilist.Current,
	}

	f, err := os.Open(MalConfigFile)
	defer f.Close()
	if os.IsNotExist(err) {
		return
	} else if err != nil {
		log.Printf("Error opening %s file: %v", MalConfigFile, err)
	}

	decoder := json.NewDecoder(f)
	decoder.Decode(c)

	return
}

func (cfg *Config) Save() {
	f, err := os.Create(MalConfigFile)
	defer f.Close()
	if err != nil {
		log.Printf("Error saving %s file: %v", MalConfigFile, err)
	}

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		log.Printf("Error encoding %s file: %v", MalConfigFile, err)
	}
}
