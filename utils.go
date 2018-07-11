package main

import (
	"encoding/base64"
	"fmt"
	"github.com/aqatl/mal/mal"
	"github.com/fatih/color"
	"log"
	"os/user"
	"time"
	"os"
	"encoding/json"
)

func basicAuth(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

func reverseAnimeSlice(s []*mal.Anime) {
	last := len(s) - 1
	for i := 0; i < len(s)/2; i++ {
		s[i], s[last-i] = s[last-i], s[i]
	}
}

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Printf("Error getting current user: %v", err)
		return ""
	}

	return usr.HomeDir
}

func printEntryDetails(entry *mal.Anime) {
	title := color.HiYellowString("%s", entry.Title)
	episodes := color.HiRedString("%d/%d", entry.WatchedEpisodes,
		entry.Episodes)
	score := color.HiRedString("%d", entry.MyScore)
	status := color.HiRedString("%v", entry.MyStatus)
	lastUpdated := color.HiRedString("%v", time.Unix(entry.LastUpdated, 0))

	fmt.Fprintf(
		color.Output,
		"Title: %s\n"+
			"Episodes: %s\n"+
			"Score: %s\n"+
			"Status: %v\n"+
			"Last updated: %v\n",
		title,
		episodes,
		score,
		status,
		lastUpdated,
	)
}

func printEntryDetailsAfterUpdatedEpisodes(entry *mal.Anime, epsBefore int) {
	title := color.HiYellowString("%s", entry.Title)
	episodesBefore := color.HiRedString("%d/%d", epsBefore,
		entry.Episodes)
	episodesAfter := color.HiRedString("%d/%d", entry.WatchedEpisodes,
		entry.Episodes)
	score := color.HiRedString("%d", entry.MyScore)
	status := color.HiRedString("%v", entry.MyStatus)

	fmt.Fprintf(
		color.Output,
		"Title: %s\n"+
			"Episodes: %s -> %s\n"+
			"Score: %s\n"+
			"Status: %v\n",
		title,
		episodesBefore,
		episodesAfter,
		score,
		status,
	)
}

//Returns true if file was loaded correctly
func LoadJsonFile(file string, i interface{}) bool {
	f, err := os.Open(file)
	defer f.Close()
	if err == nil {
		err = json.NewDecoder(f).Decode(i)
		if err == nil {
			return true
		} else {
			panic(err)
			return false
		}
	}
	if os.IsNotExist(err) {
		return false
	}
	panic(err)
	return false
}

func SaveJsonFile(file string, i interface{}) error {
	f, err := os.Create(file)
	defer f.Close()
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(i)
}
