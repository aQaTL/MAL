package main

import (
	"github.com/aqatl/mal/mal"
	"github.com/urfave/cli"
	"os"
	"log"
	"sort"
	"strconv"
	"path/filepath"
	"github.com/skratchdot/open-golang/open"
	"fmt"
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
		cli.BoolFlag{
			Name:  "creds, prompt-credentials",
			Usage: "Prompt for username and password",
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
		cli.IntFlag{
			Name:  "max",
			Usage: "visible entries threshold",
		},
		cli.StringFlag{
			Name:  "status",
			Usage: "display entries only with given status",
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
			Name:      "score",
			Category:  "Update",
			Usage:     "Set your rating for selected entry",
			UsageText: "mal score <0-10>",
			Action:    setEntryScore,
		},
		cli.Command{
			Name: "status",
			Category: "Update",
			Usage: "Set your status for selected entry",
			UsageText: "mal status [watching|completed|onhold|dropped|plantowatch]",
			Action: setEntryStatus,
		},
		cli.Command{
			Name:      "sel",
			Aliases:   []string{"select"},
			Category:  "Config",
			Usage:     "Select an entry",
			UsageText: "mal sel [entry ID]",
			Action:    selectEntry,
		},
		cli.Command{
			Name:     "cfg",
			Aliases:  []string{"config", "configuration"},
			Category: "Config",
			Usage:    "Change config values",
			Subcommands: cli.Commands{
				cli.Command{
					Name:      "max",
					Aliases:   []string{"visible"},
					Usage:     "Change amount of displayed entries",
					UsageText: "mal cfg max [number]",
					Action:    configChangeMax,
				},
				cli.Command{
					Name:      "status",
					Usage:     "Status value of displayed entries",
					UsageText: "mal cfg status [all|watching|completed|onhold|dropped|plantowatch]",
					Action:    configChangeStatus,
				},
			},
		},
		cli.Command{
			Name:      "web",
			Aliases:   []string{"website", "open"},
			Category:  "Action",
			Usage:     "Open url associated with current entry",
			UsageText: "mal web",
			Action:    openWebsite,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "url",
					Usage: "Set url for current entry",
				},
				cli.BoolFlag{
					Name:  "clear",
					Usage: "Clear url for current entry",
				},
			},
		},
	}

	app.Action = defaultAction

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func defaultAction(ctx *cli.Context) {
	creds := loadCredentials(ctx)
	if ctx.Bool("verify") && !mal.VerifyCredentials(creds) {
		log.Fatalln("Invalid credentials")
	}

	c := mal.NewClient(creds)
	if c == nil {
		os.Exit(1)
	}

	config := LoadConfig()

	list := loadList(c, ctx).FilterByStatus(statusFlag(ctx))
	sort.Sort(mal.AnimeSortByLastUpdated(list))

	var visibleEntries int
	if visibleEntries = ctx.Int("max"); visibleEntries == 0 {
		//`Max` flag not specified, get value from config
		visibleEntries = config.MaxVisibleEntries
	}
	if visibleEntries > len(list) {
		visibleEntries = len(list)
	}
	visibleList := list[:visibleEntries]
	reverseAnimeSlice(visibleList)

	PrettyList.Execute(os.Stdout, PrettyListData{visibleList, config.SelectedID})

	if ctx.GlobalBool("save-password") {
		saveCredentials(creds)
	}
}

func incrementEntry(ctx *cli.Context) error {
	creds := loadCredentials(ctx)
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
	selectedEntry := list.GetByID(cfg.SelectedID)

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

func setEntryScore(ctx *cli.Context) error {
	c := mal.NewClient(loadCredentials(ctx))
	cfg := LoadConfig()

	list := loadList(c, ctx)
	selectedEntry := list.GetByID(cfg.SelectedID)

	score, err := strconv.Atoi(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("error parsing score: %v", err)
	}
	parsedScore, err := mal.ParseScore(score)
	if err != nil {
		return err
	}

	selectedEntry.MyScore = parsedScore
	if c.Update(selectedEntry) {
		log.Printf("Updated successfully")
		cacheList(list)
	}
	return nil
}

func setEntryStatus(ctx *cli.Context) error {
	c := mal.NewClient(loadCredentials(ctx))
	cfg := LoadConfig()

	list := loadList(c, ctx)
	selectedEntry := list.GetByID(cfg.SelectedID)

	status := mal.ParseStatus(ctx.Args().First())
	if status == mal.All {
		return fmt.Errorf("invalid status; possible values: watching, completed, " +
			"onhold, dropped, plantowatch")
	}

	selectedEntry.MyStatus = status
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

func openWebsite(ctx *cli.Context) error {
	cfg := LoadConfig()

	if url := ctx.String("url"); url != "" {
		cfg.Websites[cfg.SelectedID] = url
		log.Printf("Set url %s for entry %d", url, cfg.SelectedID)
		cfg.Save()
		return nil
	}

	if ctx.Bool("clear") {
		delete(cfg.Websites, cfg.SelectedID)
		log.Printf("Cleared url for entry %d", cfg.SelectedID)
		cfg.Save()
		return nil
	}

	if url, ok := cfg.Websites[cfg.SelectedID]; ok {
		open.Start(url)
	} else {
		log.Println("Nothing to open")
	}

	return nil
}

func configChangeMax(ctx *cli.Context) error {
	cfg := LoadConfig()

	max, err := strconv.Atoi(ctx.Args().First())
	if err != nil || max < 0 {
		log.Fatalf("Invalid value")
	}

	cfg.MaxVisibleEntries = max
	cfg.Save()
	return nil
}

func configChangeStatus(ctx *cli.Context) error {
	cfg := LoadConfig()

	status := mal.ParseStatus(ctx.Args().First())

	cfg.Status = status
	cfg.Save()
	return nil
}
