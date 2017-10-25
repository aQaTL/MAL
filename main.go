package main

import (
	"github.com/aqatl/mal/mal"
	"github.com/urfave/cli"
	"os"
	"log"
	"sort"
	"strconv"
	"path/filepath"
)

var dataDir = filepath.Join(homeDir(), ".mal")
var (
	CredentialsFile = filepath.Join(dataDir, "cred.dat")
	MalCacheFile    = filepath.Join(dataDir, "cache.xml")
	ConfigFile      = filepath.Join(dataDir, "config.json")
)

func main() {
	checkDataDir()

	app := cli.NewApp()
	app.Name = "mal"
	app.Usage = "App for managing your MAL"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "u, username",
			Usage: "specify username",
		},
		cli.StringFlag{
			Name:  "p, password",
			Usage: "specify password",
		},
		cli.BoolFlag{
			Name:  "sp, save-password",
			Usage: "save your password. Use with caution, your password can be decoded",
		},
		cli.BoolFlag{
			Name:  "r, refresh",
			Usage: "refreshes cache file",
		},
		cli.BoolFlag{
			Name:  "ver, verify",
			Usage: "verify credentials",
		},
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:      "inc",
			Aliases:   []string{"+1"},
			Category:  "Update",
			Usage:     "Increment selected entry by one",
			UsageText: "mal inc",
			Action:    incrementEntry,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "n",
					Usage: "Specify exact episode to set the entry to",
				},
			},
		},
		cli.Command{
			Name:      "sel",
			Aliases:   []string{"select"},
			Category:  "Config",
			Usage:     "Select an entry",
			UsageText: "mal sel [entry ID]",
			Action:    selectEntry,
		},
	}

	app.Action = defaultAction

	if err := app.Run(os.Args); err != nil {
		log.Printf("Arguments error: %v", err)
		os.Exit(1)
	}
}

func defaultAction(ctx *cli.Context) {
	creds := credentials(ctx)
	if ctx.Bool("verify") && !mal.VerifyCredentials(creds) {
		log.Fatalln("Invalid credentials")
	}

	c := mal.NewClient(creds)
	if c == nil {
		os.Exit(1)
	}

	config := LoadConfig()

	list := loadList(c, ctx)
	sort.Sort(mal.AnimeSortByLastUpdated(list))

	visibleEntries := config.MaxVisibleEntries
	if visibleEntries > len(list) {
		visibleEntries = len(list)
	}
	visibleList := list[:visibleEntries]
	reverseAnimeSlice(visibleList)

	PrettyList.Execute(os.Stdout, PrettyListData{visibleList, config.SelectedID})

	if ctx.Bool("save-password") {
		cacheCredentials(ctx.String("username"), ctx.String("password"))
	}

	cacheList(list)
}

func incrementEntry(ctx *cli.Context) error {
	creds := credentials(ctx)
	if ctx.Bool("verify") && !mal.VerifyCredentials(creds) {
		log.Fatalln("Invalid credentials")
	}

	c := mal.NewClient(creds)
	if c == nil {
		os.Exit(1)
	}
	cfg := LoadConfig()

	if cfg.SelectedID == 0 {
		log.Fatalln("No entry selected")
	}

	list := loadList(c, ctx)
	var selectedEntry *mal.Anime
	for i, entry := range list {
		if entry.ID == cfg.SelectedID {
			selectedEntry = list[i]
			break
		}
	}

	if selectedEntry == nil {
		log.Fatalln("No entry found")
	}

	if ctx.Int("n") > 0 {
		selectedEntry.WatchedEpisodes = ctx.Int("n")
	} else {
		selectedEntry.WatchedEpisodes++
	}

	if c.Update(selectedEntry) {
		log.Printf("Updated successfully")
		cacheList(list)
	}
	return nil
}

func selectEntry(ctx *cli.Context) error {
	cfg := LoadConfig()

	id, err := strconv.Atoi(ctx.Args().First())
	if err != nil {
		log.Fatalf("Error parsing id: %v", err)
	}

	cfg.SelectedID = id
	cfg.Save()
	return nil
}
