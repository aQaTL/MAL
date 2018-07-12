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
	"github.com/aqatl/mal/anilist"
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

func chooseStrFromSlice(alts []string) string {
	if length := len(alts); length == 1 {
		return alts[0]
	} else if length == 0 {
		return ""
	}

	for i, synonym := range alts {
		fmt.Printf("%2d. %s\n", i+1, synonym)
	}

	idx := 0
	scan := func() {
		fmt.Scan(&idx)
	}
	for scan(); idx <= 0 || idx > len(alts); {
		fmt.Print("\rInvalid input. Try again: ")
		scan()
	}

	return alts[idx-1]
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

func alPrintEntryDetails(entry *anilist.MediaList) {
	title := color.HiYellowString("%s", entry.Title.UserPreferred)
	episodes := color.HiRedString("%d/%d", entry.Progress,
		entry.Episodes)
	score := color.HiRedString("%d", entry.Score)
	status := color.HiRedString("%v", entry.Status)
	lastUpdated := color.HiRedString("%v", time.Unix(int64(entry.UpdatedAt), 0))

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
