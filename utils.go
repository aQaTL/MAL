package main

import (
	"encoding/base64"
	"fmt"
	"github.com/aqatl/mal/mal"
	"github.com/fatih/color"
	"log"
	"os/user"
	"time"
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

	fmt.Printf(
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

	fmt.Printf(
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
