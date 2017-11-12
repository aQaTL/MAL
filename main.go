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
	"strings"
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
	app.Version = "0.1"

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
			Name:      "status",
			Category:  "Update",
			Usage:     "Set your status for selected entry",
			UsageText: "mal status [watching|completed|onhold|dropped|plantowatch]",
			Action:    setEntryStatus,
		},
		cli.Command{
			Name:      "sel",
			Aliases:   []string{"select"},
			Category:  "Config",
			Usage:     "Select an entry",
			UsageText: "mal sel [entry ID]",
			Action:    selectEntry,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "by-title",
					Usage: "Select entry by name instead of by ID",
				},
			},
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
				cli.Command{
					Name:      "status-auto-update",
					Usage:     "Allows entry to be automatically set to completed when number of all episodes is reached or exceeded",
					UsageText: "mal cfg status-auto-update [off|normal|after-threshold]",
					Action:    configChangeAutoUpdateMode,
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
			Subcommands: []cli.Command{
				cli.Command{
					Name:      "get-all",
					Usage:     "Print all set urls",
					UsageText: "mal web get-all",
					Action:    printWebsites,
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
	if ctx.GlobalBool("verify") && !mal.VerifyCredentials(creds) {
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
		return fmt.Errorf("invalid credentials")
	}

	c := mal.NewClient(creds)
	if c == nil {
		os.Exit(1)
	}
	cfg := LoadConfig()

	if cfg.SelectedID == 0 {
		return fmt.Errorf("no entry selected")
	}

	list := loadList(c, ctx)
	selectedEntry := list.GetByID(cfg.SelectedID)

	if selectedEntry == nil {
		return fmt.Errorf("no entry found")
	}

	if ctx.Int("n") > 0 {
		selectedEntry.WatchedEpisodes = ctx.Int("n")
	} else {
		selectedEntry.WatchedEpisodes++
	}

	statusAutoUpdate(cfg, selectedEntry)

	if selectedEntry.WatchedEpisodes > selectedEntry.Episodes {
		selectedEntry.WatchedEpisodes = selectedEntry.Episodes
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
	if ctx.Bool("by-title") {
		return selectByTitle(ctx)
	}

	id, err := strconv.Atoi(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("error parsing id: %v", err)
	}

	cfg := LoadConfig()
	cfg.SelectedID = id
	cfg.Save()
	return nil
}

func selectByTitle(ctx *cli.Context) error {
	title := strings.ToLower(ctx.Args().First())
	list := loadList(mal.NewClient(loadCredentials(ctx)), ctx)

	found := make(mal.AnimeList, 0)
	for _, entry := range list {
		if strings.Contains(strings.ToLower(entry.Title), title) {
			found = append(found, entry)
		}
	}

	var id int

	if len(found) > 1 {
		fmt.Printf("Found more than 1 matching entry:\n")
		fmt.Printf("%3s%8s%7s\n", "No.", "ID", "Title")
		fmt.Println(strings.Repeat("=", 80))
		for i, entry := range found {
			fmt.Printf("%3d. %6d: %s\n", i, entry.ID, entry.Title)
		}

		fmt.Printf("Enter index of the selected entry: ")
		idx := 0
		_, err := fmt.Scanln(&idx)
		if err != nil || idx < 0 || idx > len(found)-1 {
			return fmt.Errorf("invalid input %v", err)
		}

		id = found[idx].ID
	} else if len(found) == 0 {
		return fmt.Errorf("no matches")
	} else {
		id = found[0].ID
	}

	cfg := LoadConfig()
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

func printWebsites(ctx *cli.Context) error {
	cfg := LoadConfig()
	list := loadList(mal.NewClient(loadCredentials(ctx)), ctx)

	for k, v := range cfg.Websites {
		url := fmt.Sprintf("\033[3%d;%dm%s\033[0m ", 3, 1, v)

		var title string
		if entry := list.GetByID(k); entry != nil {
			title = entry.Title
		}

		fmt.Printf("%6d (%s): %s\n", k, title, url)
	}

	return nil
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

func configChangeStatus(ctx *cli.Context) error {
	cfg := LoadConfig()

	status := mal.ParseStatus(ctx.Args().First())

	cfg.Status = status
	cfg.Save()
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
