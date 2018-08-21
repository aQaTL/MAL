package main

import (
	"encoding/json"
	"fmt"
	"github.com/aqatl/mal/anilist"
	"github.com/aqatl/mal/mal"
	"github.com/fatih/color"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"github.com/urfave/cli"
)

type Config struct {
	Websites map[int]string

	MaxVisibleEntries    int
	StatusAutoUpdateMode StatusAutoUpdateMode
	Sorting              Sorting
	LastUpdate           time.Time

	BrowserPath       string
	TorrentClientPath string
	TorrentClientArgs []string

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
		Websites: make(map[int]string),

		MaxVisibleEntries:    20,
		StatusAutoUpdateMode: Off,
		Sorting:              ByLastUpdated,

		TorrentClientPath: "qbittorrent",

		Status: mal.All,

		ALStatus: anilist.Current,
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

func configChangeMax(ctx *cli.Context) error {
	cfg := LoadConfig()

	max, err := strconv.Atoi(ctx.Args().First())
	if err != nil || max < 0 {
		return fmt.Errorf("invalid value")
	}

	cfg.MaxVisibleEntries = max
	cfg.Save()
	return nil
}

func configChangeMalStatus(ctx *cli.Context) error {
	cfg := LoadConfig()

	status := mal.ParseStatus(ctx.Args().First())

	cfg.Status = status
	cfg.Save()
	return nil
}

func configChangeAlStatus(ctx *cli.Context) error {
	cfg := LoadConfig()

	status := anilist.ParseStatus(ctx.Args().First())

	cfg.ALStatus = status
	cfg.Save()

	str := status.String()
	if status == anilist.All {
		str = "All"
	}
	fmt.Println("New status:", str)
	return nil
}

func configChangeAutoUpdateMode(ctx *cli.Context) error {
	arg := strings.ToLower(ctx.Args().First())
	var mode StatusAutoUpdateMode

	if arg == "off" {
		mode = Off
	} else if arg == "normal" {
		mode = Normal
	} else if arg == "after-threshold" {
		mode = AfterThreshold
	} else {
		return fmt.Errorf("invalid option; possible values: off|normal|after-threshold")
	}

	cfg := LoadConfig()
	cfg.StatusAutoUpdateMode = mode
	cfg.Save()

	return nil
}

func configChangeSorting(ctx *cli.Context) error {
	sorting, err := ParseSorting(strings.ToLower(ctx.Args().First()))
	if err != nil {
		return fmt.Errorf("error parsing flags: %v", err)
	}

	cfg := LoadConfig()
	cfg.Sorting = sorting
	cfg.Save()

	return nil
}

func configChangeBrowser(ctx *cli.Context) error {
	cfg := LoadConfig()

	if ctx.Bool("clear") {
		cfg.BrowserPath = ""
		cfg.Save()

		fmt.Printf("Browser path cleared\n")
		return nil
	}

	browserPath, err := filepath.Abs(strings.Join(ctx.Args(), " "))
	if err != nil {
		return fmt.Errorf("path error: %v", err)
	}

	cfg.BrowserPath = browserPath
	cfg.Save()

	fmt.Fprintf(color.Output, "New browser path: %v\n", color.HiYellowString("%v", browserPath))

	return nil
}

func configChangeTorrent(ctx *cli.Context) error {
	cfg := LoadConfig()

	cfg.TorrentClientPath = ctx.Args().First()
	cfg.TorrentClientArgs = ctx.Args().Tail()

	cfg.Save()

	fmt.Fprintf(
		color.Output,
		"New torrent config: %s %s\n",
		color.HiYellowString("%s", cfg.TorrentClientPath),
		color.HiCyanString("%s", strings.Join(cfg.TorrentClientArgs, " ")))

	return nil
}
