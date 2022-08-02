package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/aqatl/mal/anilist"
	"github.com/aqatl/mal/mal"
	"github.com/fatih/color"
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

func getDataDir() string {
	// Check for old cache dir at $HOME/.mal
	if usr, err := user.Current(); err == nil {
		dir := filepath.Join(usr.HomeDir, ".mal")
		if _, err := os.Stat(dir); err == nil {
			return dir
		} else {
			if !os.IsNotExist(err) {
				log.Printf("Error probing for %s: %v", dir, err)
			}
		}
	} else {
		log.Printf("Error getting current user: %v. ignoring", err)
	}

	// Old cache dir not present, use user config dir
	dataDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Error getting user config dir: %v", err)
		return ""
	}

	return filepath.Join(dataDir, "mal")
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

func printEntryDetails(title, status string, watchedEps, eps, score int, lastUpdated time.Time) {
	titleStr := color.HiYellowString("%s", title)
	episodesStr := color.HiRedString("%d/%d", watchedEps, eps)
	scoreStr := color.HiRedString("%d", score)
	statusStr := color.HiRedString("%s", status)
	lastUpdatedStr := color.HiRedString("%v", lastUpdated)

	fmt.Fprintf(
		color.Output,
		"Title: %s\n"+
			"Episodes: %s\n"+
			"Score: %s\n"+
			"Status: %v\n"+
			"Last updated: %v\n",
		titleStr,
		episodesStr,
		scoreStr,
		statusStr,
		lastUpdatedStr,
	)
}

func printEntryDetailsAfterUpdatedEpisodes(title, status string, epsBefore, epsNow, eps, score int, lastUpdated time.Time) {
	titleStr := color.HiYellowString("%s", title)
	episodesBeforeStr := color.HiRedString("%d/%d", epsBefore, eps)
	episodesAfterStr := color.HiRedString("%d/%d", epsNow, eps)
	scoreStr := color.HiRedString("%d", score)
	statusStr := color.HiRedString("%s", status)
	lastUpdatedStr := color.HiRedString("%v", lastUpdated)

	fmt.Fprintf(
		color.Output,
		"Title: %s\n"+
			"Episodes: %s -> %s\n"+
			"Score: %s\n"+
			"Status: %v\n"+
			"Last updated: %v\n",
		titleStr,
		episodesBeforeStr,
		episodesAfterStr,
		scoreStr,
		statusStr,
		lastUpdatedStr,
	)
}

func malPrintEntryDetails(entry *mal.Anime) {
	printEntryDetails(
		entry.Title,
		entry.MyStatus.String(),
		entry.WatchedEpisodes,
		entry.Episodes,
		int(entry.MyScore),
		time.Unix(entry.LastUpdated, 0))
}

func malPrintEntryDetailsAfterUpdatedEpisodes(entry *mal.Anime, epsBefore int) {
	printEntryDetailsAfterUpdatedEpisodes(
		entry.Title,
		entry.MyStatus.String(),
		epsBefore,
		entry.WatchedEpisodes,
		entry.Episodes,
		int(entry.MyScore),
		time.Unix(entry.LastUpdated, 0))
}

func alPrintEntryDetails(entry *anilist.MediaListEntry) {
	printEntryDetails(entry.Title.UserPreferred,
		entry.Status.String(),
		entry.Progress,
		entry.Episodes,
		entry.Score,
		time.Unix(int64(entry.UpdatedAt), 0))
}

func alPrintEntryDetailsAfterUpdatedEpisodes(entry *anilist.MediaListEntry, epsBefore int) {
	printEntryDetailsAfterUpdatedEpisodes(
		entry.Title.UserPreferred,
		entry.Status.String(),
		epsBefore,
		entry.Progress,
		entry.Episodes,
		entry.Score,
		time.Unix(int64(entry.UpdatedAt), 0))
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
