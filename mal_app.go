package main

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aqatl/cliwait"
	"github.com/aqatl/mal/mal"
	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli"
)

func MalApp(app *cli.App) *cli.App {
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
		cli.BoolFlag{
			Name:  "all, a",
			Usage: "display all entries; same as --max -1",
		},
		cli.StringFlag{
			Name: "status",
			Usage: "display entries only with given status " +
				"[watching|completed|onhold|dropped|plantowatch]",
		},
		cli.StringFlag{
			Name:  "sort",
			Usage: "display entries sorted by: [last-updated|title|episodes|score]",
		},
		cli.BoolFlag{
			Name:  "reversed",
			Usage: "reversed list order",
		},
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:      "switch",
			Aliases:   []string{"s"},
			Usage:     "Switches app mode to AniList",
			UsageText: "mal anilist",
			Action:    switchToAniList,
		},
		cli.Command{
			Name:     "eps",
			Aliases:  []string{"episodes"},
			Category: "Update",
			Usage: "Set the watched episodes value. " +
				"If n not specified, the number will be increased by one",
			UsageText: "mal eps <n>",
			Action:    setEntryEpisodes,
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
			Name:      "cmpl",
			Category:  "Update",
			Usage:     "Set entry status to completed",
			UsageText: "mal cmpl",
			Action:    setEntryStatusCompleted,
		},
		cli.Command{
			Name:      "delete",
			Aliases:   []string{"del"},
			Category:  "Update",
			Usage:     "Delete entry from your list",
			UsageText: "mal delete",
			Action:    deleteEntry,
		},
		cli.Command{
			Name:      "sel",
			Aliases:   []string{"select"},
			Category:  "Config",
			Usage:     "Select an entry",
			UsageText: "mal sel [entry title]",
			Action:    selectEntry,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "id",
					Usage: "Select entry by id instead of by title",
				},
			},
		},
		cli.Command{
			Name:      "fuzzy-select",
			Aliases:   []string{"fsel"},
			Category:  "Config",
			Usage:     "Interactive fuzzy search through your list",
			UsageText: "mal fsel [search string (optional)]",
			Action:    fuzzySelectEntry,
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
					Action:    configChangeMalStatus,
				},
				cli.Command{
					Name:      "status-auto-update",
					Usage:     "Allows entry to be automatically set to completed when number of all episodes is reached or exceeded",
					UsageText: "mal cfg status-auto-update [off|normal|after-threshold]",
					Action:    configChangeAutoUpdateMode,
				},
				cli.Command{
					Name:      "sort",
					Usage:     "Specifies sorting mode for the displayed table",
					UsageText: "mal cfg sort [last-updated|title|progress|score]",
					Action:    configChangeSorting,
				},
				cli.Command{
					Name:      "browser",
					Usage:     "Specifies a browser to use",
					UsageText: "mal cfg browser [browser_path]",
					Action:    configChangeBrowser,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "clear",
							Usage: "Clear browser path (return to default)",
						},
					},
				},
				cli.Command{
					Name:            "torrent",
					Usage:           "Sets path to torrent client and it args",
					UsageText:       "mal cfg torrent [path] [args...]",
					SkipFlagParsing: true,
					Action:          configChangeTorrent,
				},
				cli.Command{
					Name:            "nyaa-quality",
					Usage:           "Sets default quality filter for nyaa search",
					UsageText:       "mal cfg nyaa-quality [quality_text]",
					SkipFlagParsing: true,
					Action:          configChangeNyaaQuality,
				},
			},
		},
		cli.Command{
			Name:      "web",
			Aliases:   []string{"website", "open", "url"},
			Category:  "Action",
			Usage:     "Open url associated with selected entry or change url if provided",
			UsageText: "mal web <url>",
			Action:    openWebsite,
			Flags: []cli.Flag{
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
		cli.Command{
			Name:      "stats",
			Category:  "Action",
			Usage:     "Show your account statistics",
			UsageText: "mal stats",
			Action:    malStats,
		},
		cli.Command{
			Name:      "mal",
			Category:  "Action",
			Usage:     "Open MyAnimeList page of selected entry",
			UsageText: "mal mal",
			Action:    malOpenMalSite,
		},
		cli.Command{
			Name:      "details",
			Category:  "Action",
			Usage:     "Print details about selected entry",
			UsageText: "mal details",
			Action:    printDetails,
		},
		cli.Command{
			Name:      "related",
			Category:  "Action",
			Usage:     "Fetch entries related to the selected one",
			UsageText: "mal related",
			Action:    printRelated,
		},
		cli.Command{
			Name:      "music",
			Category:  "Action",
			Usage:     "Print opening and ending themes",
			UsageText: "mal music",
			Action:    printMusic,
		},
		cli.Command{
			Name:      "broadcast",
			Category:  "Action",
			Usage:     "Print broadcast (airing) time",
			UsageText: "mal broadcast",
			Action:    printBroadcast,
		},
		cli.Command{
			Name:      "copy",
			Category:  "Action",
			Usage:     "Copy selected value into system clipboard",
			UsageText: "mal copy [title|url]",
			Action:    copyIntoClipboard,
		},
		cli.Command{
			Name:     "nyaa",
			Aliases:  []string{"n"},
			Category: "Action",
			Usage:    "Open interactive torrent search",
			Action:   malNyaaCui,
		},
		cli.Command{
			Name:      "nyaa-web",
			Aliases:   []string{"nw"},
			Category:  "Action",
			Usage:     "Open torrent search in browser",
			UsageText: "mal nyaa-web",
			Action:    nyaaWebsite,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "alt",
					Usage: "choose an alternative title",
				},
			},
		},
	}

	app.Action = cli.ActionFunc(malDefaultAction)

	return app
}

func loadMAL(ctx *cli.Context) (*mal.Client, mal.AnimeList, error) {
	creds := loadCredentials(ctx)
	if ctx.GlobalBool("verify") && !mal.VerifyCredentials(creds) {
		return nil, nil, fmt.Errorf("invalid credentials")
	}
	if ctx.GlobalBool("save-password") {
		saveCredentials(creds)
	}

	c := mal.NewClient(creds)
	if c == nil {
		return nil, nil, fmt.Errorf("error creating mal.Client")
	}

	list, err := loadData(c, ctx)

	return c, list, err
}

func malDefaultAction(ctx *cli.Context) error {
	_, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}

	cfg := LoadConfig()

	list = list.FilterByStatus(statusFlag(ctx))

	sorting := cfg.Sorting
	if ctx.String("sort") != "" {
		if sorting, err = ParseSorting(ctx.String("sort")); err != nil {
			return fmt.Errorf("error parsing 'sort' option: %v", err)
		}
	}

	switch sorting {
	case ByLastUpdated:
		sort.Sort(mal.AnimeSortByLastUpdated(list))
	case ByTitle:
		sort.Sort(mal.AnimeSortByTitle(list))
	case ByWatchedEpisodes:
		sort.Sort(sort.Reverse(mal.AnimeSortByWatchedEpisodes(list)))
	case ByScore:
		sort.Sort(sort.Reverse(mal.AnimeSortByScore(list)))
	default:
		sort.Sort(mal.AnimeSortByLastUpdated(list))
	}

	if ctx.Bool("reversed") {
		reverseAnimeSlice(list)
	}
	var visibleEntries int
	if visibleEntries = ctx.Int("max"); visibleEntries == 0 {
		//`Max` flag not specified, get value from config
		visibleEntries = cfg.MaxVisibleEntries
	}
	if visibleEntries > len(list) || visibleEntries < 0 || ctx.Bool("all") {
		visibleEntries = len(list)
	}
	visibleList := list[:visibleEntries]
	reverseAnimeSlice(visibleList)

	PrettyList.Execute(color.Output, PrettyListData{visibleList, cfg.SelectedID})

	if cfg.LastUpdate != *new(time.Time) {
		fmt.Printf("\nList last updated: %v (%d days ago)\n",
			cfg.LastUpdate,
			int(time.Since(cfg.LastUpdate).Hours()/24),
		)
	}

	return nil
}

func statusFlag(ctx *cli.Context) mal.MyStatus {
	var status mal.MyStatus
	if customStatus := ctx.GlobalString("status"); customStatus != "" {
		status = mal.ParseStatus(customStatus)
	} else {
		cfg := LoadConfig()
		status = cfg.Status
	}
	return status
}

func statusAutoUpdate(cfg *Config, entry *mal.Anime) {
	if cfg.StatusAutoUpdateMode == Off || entry.Episodes == 0 {
		return
	}

	if (cfg.StatusAutoUpdateMode == Normal && entry.WatchedEpisodes >= entry.Episodes) ||
		(cfg.StatusAutoUpdateMode == AfterThreshold && entry.WatchedEpisodes > entry.Episodes) {
		entry.MyStatus = mal.Completed
		entry.WatchedEpisodes = entry.Episodes
		return
	}

	if entry.MyStatus == mal.Completed && entry.WatchedEpisodes < entry.Episodes {
		entry.MyStatus = mal.Watching
		return
	}
}

func switchToAniList(ctx *cli.Context) error {
	appCfg := AppConfig{}
	LoadJsonFile(AppConfigFile, &appCfg)
	appCfg.Mode = AniListMode
	if err := SaveJsonFile(AppConfigFile, &appCfg); err != nil {
		return nil
	}
	fmt.Println("App mode switched to AniList")
	return nil
}

func setEntryEpisodes(ctx *cli.Context) error {
	c, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}

	cfg := LoadConfig()

	if cfg.SelectedID == 0 {
		return fmt.Errorf("no entry selected")
	}

	selectedEntry := list.GetByID(cfg.SelectedID)

	if selectedEntry == nil {
		return fmt.Errorf("no entry found")
	}

	epsBefore := selectedEntry.WatchedEpisodes

	if arg := ctx.Args().First(); arg != "" {
		n, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("n must be a non-negative integer")
		}
		if n < 0 {
			return fmt.Errorf("n can't be lower than 0")
		}
		selectedEntry.WatchedEpisodes = n
	} else {
		selectedEntry.WatchedEpisodes++
	}

	statusAutoUpdate(cfg, selectedEntry)

	if err := mal.UpdateEntryWithAnimation(c, selectedEntry); err != nil {
		return err
	}

	fmt.Println("Updated successfully")
	malPrintEntryDetailsAfterUpdatedEpisodes(selectedEntry, epsBefore)

	cacheList(list)

	return nil
}

func setEntryScore(ctx *cli.Context) error {
	c, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}
	cfg := LoadConfig()

	selectedEntry := list.GetByID(cfg.SelectedID)
	if selectedEntry == nil {
		return fmt.Errorf("no entry selected")
	}

	score, err := strconv.Atoi(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("error parsing score: %v", err)
	}
	parsedScore, err := mal.ParseScore(score)
	if err != nil {
		return err
	}

	selectedEntry.MyScore = parsedScore

	if err := mal.UpdateEntryWithAnimation(c, selectedEntry); err != nil {
		return err
	}

	fmt.Println("Updated successfully")
	malPrintEntryDetails(selectedEntry)

	cacheList(list)

	return nil
}

func setEntryStatus(ctx *cli.Context) error {
	c, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}
	cfg := LoadConfig()

	selectedEntry := list.GetByID(cfg.SelectedID)
	if selectedEntry == nil {
		return fmt.Errorf("no entry selected")
	}

	status := mal.ParseStatus(ctx.Args().First())
	if status == mal.All {
		return fmt.Errorf("invalid status; possible values: watching, completed, " +
			"onhold, dropped, plantowatch")
	}

	selectedEntry.MyStatus = status

	if err := mal.UpdateEntryWithAnimation(c, selectedEntry); err != nil {
		return err
	}

	fmt.Println("Updated successfully")
	malPrintEntryDetails(selectedEntry)

	cacheList(list)

	return nil
}

func setEntryStatusCompleted(ctx *cli.Context) error {
	c, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}
	cfg := LoadConfig()

	selectedEntry := list.GetByID(cfg.SelectedID)
	if selectedEntry == nil {
		return fmt.Errorf("no entry selected")
	}

	selectedEntry.MyStatus = mal.Completed
	selectedEntry.WatchedEpisodes = selectedEntry.Episodes

	if err := mal.UpdateEntryWithAnimation(c, selectedEntry); err != nil {
		return err
	}

	fmt.Println("Updated successfully")
	malPrintEntryDetails(selectedEntry)

	cacheList(list)

	return nil
}

func deleteEntry(ctx *cli.Context) error {
	c, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}
	cfg := LoadConfig()

	entry := list.GetByID(cfg.SelectedID)
	if entry == nil {
		return fmt.Errorf("no entry selected")
	}

	cliwait.DoFuncWithWaitAnimation("Deleting entry", func() {
		err = c.Delete(entry)
	})
	if err != nil {
		return fmt.Errorf("deleting entry failed\n%v", err)
	}

	title := color.HiRedString("%s", entry.Title)
	fmt.Fprintf(color.Output, "%s seleted successfully\n", title)
	list = list.DeleteByID(entry.ID)
	cacheList(list)

	return nil
}

func selectEntry(ctx *cli.Context) error {
	switch {
	case ctx.Bool("id"):
		return selectById(ctx)
	default:
		return selectByTitle(ctx)
	}
}

func selectById(ctx *cli.Context) error {
	id, err := strconv.Atoi(ctx.Args().First())
	if err != nil {
		return fmt.Errorf("invalid id (use with -t to select by title)")
	}

	_, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}
	cfg := LoadConfig()

	entry := list.GetByID(id)
	if entry == nil {
		return fmt.Errorf("entry %d not found", id)
	}

	cfg.SelectedID = id
	cfg.Save()

	fmt.Println("Selected entry:")
	malPrintEntryDetails(entry)

	return nil
}

func selectByTitle(ctx *cli.Context) error {
	title := strings.ToLower(strings.Join(ctx.Args(), " "))
	if title == "" {
		return showSelectedEntry(ctx)
	}

	_, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}

	found := make(mal.AnimeList, 0)
	for _, entry := range list {
		if strings.Contains(strings.ToLower(entry.Title), title) ||
			strings.Contains(strings.ToLower(entry.Synonyms), title) {
			found = append(found, entry)
		}
	}

	var selectedEntry *mal.Anime

	if len(found) > 1 {
		fmt.Printf("Found more than 1 matching entry:\n")
		fmt.Printf("%3s%8s%7s\n", "No.", "ID", "Title")
		fmt.Println(strings.Repeat("=", 80))

		sort.Sort(mal.AnimeSortByLastUpdated(found))
		for i, entry := range found {
			fmt.Printf("%3d. %6d: %s\n", i+1, entry.ID, entry.Title)
		}

		fmt.Printf("Enter index of the selected entry: ")
		idx := 0
		_, err := fmt.Scanln(&idx)
		idx-- //List is displayed from 1
		if err != nil || idx < 0 || idx > len(found)-1 {
			return fmt.Errorf("invalid input %v", err)
		}

		selectedEntry = found[idx]
	} else if len(found) == 0 {
		return fmt.Errorf("no matches")
	} else {
		selectedEntry = found[0]
	}

	cfg := LoadConfig()
	cfg.SelectedID = selectedEntry.ID
	cfg.Save()

	fmt.Println("Selected entry:")
	malPrintEntryDetails(selectedEntry)

	return nil
}

func showSelectedEntry(ctx *cli.Context) error {
	cfg := LoadConfig()
	_, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}

	selEntry := list.GetByID(cfg.SelectedID)
	if selEntry == nil {
		return fmt.Errorf("no entry selected")
	}

	fmt.Println("Selected entry:")
	malPrintEntryDetails(selEntry)

	return nil
}

func openWebsite(ctx *cli.Context) error {
	_, list, err := loadMAL(ctx)
	if err != nil {
		return nil
	}

	cfg := LoadConfig()

	entry := list.GetByID(cfg.SelectedID)
	if entry == nil {
		return fmt.Errorf("no entry selected")
	}

	if newUrl := ctx.Args().First(); newUrl != "" {
		cfg.Websites[cfg.SelectedID] = newUrl
		cfg.Save()

		fmt.Print("Entry: ")
		color.HiYellow("%v", entry.Title)
		fmt.Print("URL: ")
		color.HiRed("%v", cfg.Websites[cfg.SelectedID])

		return nil
	}

	if ctx.Bool("clear") {
		delete(cfg.Websites, cfg.SelectedID)
		cfg.Save()

		fmt.Println("Entry cleared")
		return nil
	}

	if entryUrl, ok := cfg.Websites[cfg.SelectedID]; ok {
		if path := cfg.BrowserPath; path == "" {
			open.Start(entryUrl)
		} else {
			open.StartWith(entryUrl, path)
		}

		fmt.Println("Opened website for:")
		malPrintEntryDetails(entry)
		fmt.Fprintf(color.Output, "URL: %v\n", color.CyanString("%v", entryUrl))
	} else {
		fmt.Println("Nothing to open")
	}

	return nil
}

func nyaaWebsite(ctx *cli.Context) error {
	cfg := LoadConfig()
	_, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}

	entry := list.GetByID(cfg.SelectedID)
	if entry == nil {
		return fmt.Errorf("no entry selected")
	}

	var searchTerm string
	if ctx.Bool("alt") {
		synonyms := formatSynonyms(entry.Synonyms, func(a ...interface{}) string {
			return a[0].(string)
		})
		if searchTerm = chooseStrFromSlice(synonyms); searchTerm == "" {
			return fmt.Errorf("no alternative titles")
		}
	} else {
		searchTerm = entry.Title
	}

	address := "https://nyaa.si/?f=0&c=1_2&q=" + url.QueryEscape(searchTerm)

	if path := cfg.BrowserPath; path == "" {
		open.Start(address)
	} else {
		open.StartWith(address, path)
	}

	fmt.Println("Searched for:")
	malPrintEntryDetails(entry)

	return nil
}

func malStats(ctx *cli.Context) error {
	c, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}

	yellow := color.New(color.FgHiYellow).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()
	cyan := color.New(color.FgHiCyan).SprintfFunc()
	magenta := color.New(color.FgHiMagenta).SprintFunc()

	totalEntries := c.Watching + c.Completed + c.Dropped + c.OnHold + c.PlanToWatch

	watchedEps, rewatchedSeries := 0, 0
	for _, entry := range list {
		watchedEps += entry.WatchedEpisodes
		rewatchedSeries += entry.MyRewatching
	}

	hoursSpentWatching := c.DaysSpentWatching * 24.0

	fmt.Fprintf(
		color.Output,
		"Username: %s\n\n"+
			"Watching: %s\n"+
			"Completed: %s\n"+
			"Dropped: %s\n"+
			"On hold: %s\n"+
			"Plan to watch: %s\n\n"+
			"Total entries: %s\n"+
			"Episodes watched: %s\n"+
			"Times rewatched: %s\n\n"+
			"Days spent watching: %s (%s hours)\n",
		yellow(c.Username),
		red(c.Watching),
		red(c.Completed),
		red(c.Dropped),
		red(c.OnHold),
		red(c.PlanToWatch),
		magenta(totalEntries),
		magenta(watchedEps),
		magenta(rewatchedSeries),
		cyan("%.2f", c.DaysSpentWatching),
		cyan("%.2f", hoursSpentWatching),
	)

	return nil
}

func malOpenMalSite(ctx *cli.Context) error {
	_, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}

	cfg := LoadConfig()
	id := cfg.SelectedID
	if id <= 0 {
		return fmt.Errorf("no entry selected")
	}

	openMalSite(cfg, cfg.SelectedID)
	fmt.Println("Opened website for:")
	malPrintEntryDetails(list.GetByID(cfg.SelectedID))

	return nil
}

func openMalSite(cfg *Config, malId int) {
	if path, args := cfg.BrowserPath, fmt.Sprintf(mal.AnimePage, malId); path == "" {
		open.Start(args)
	} else {
		open.StartWith(args, path)
	}
}

func printWebsites(ctx *cli.Context) error {
	_, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}

	cfg := LoadConfig()

	for k, v := range cfg.Websites {
		entryUrl := fmt.Sprintf("\033[3%d;%dm%s\033[0m ", 3, 1, v)

		var title string
		if entry := list.GetByID(k); entry != nil {
			title = entry.Title
		}

		fmt.Fprintf(color.Output, "%6d (%s): %s\n", k, title, entryUrl)
	}

	return nil
}

func printDetails(ctx *cli.Context) error {
	cfg := LoadConfig()
	c, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}

	entry := list.GetByID(cfg.SelectedID)
	if entry == nil {
		return fmt.Errorf("no entry selected")
	}

	details, err := mal.FetchDetailsWithAnimation(c, entry)
	if err != nil {
		return err
	}

	printSlice := func(slice []string) {
		for _, str := range slice {
			fmt.Fprintf(color.Output, "\t%s\n", str)
		}
	}

	yellow := color.New(color.FgHiYellow).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()
	cyan := color.New(color.FgHiCyan).SprintFunc()
	green := color.New(color.FgHiGreen).SprintFunc()

	fmt.Fprintln(color.Output, "Title:", yellow(entry.Title))
	fmt.Fprintln(color.Output, "Japanese title:", yellow(details.JapaneseTitle))
	fmt.Fprintln(color.Output, "Series synonyms:")
	printSlice(formatSynonyms(entry.Synonyms, yellow))
	fmt.Fprintln(color.Output, "Series type:", yellow(entry.Type))
	fmt.Fprintln(color.Output, "Series status:", yellow(entry.Status))
	fmt.Fprintln(color.Output, "Series premiered:", yellow(details.Premiered))
	fmt.Fprintln(color.Output, "Series start:", yellow(entry.SeriesStart))
	fmt.Fprintln(color.Output, "Series end:", yellow(entry.SeriesEnd))
	fmt.Fprintln(color.Output, "Series score:", red(details.Score),
		"(by", red(details.ScoreVoters), "voters)")
	fmt.Fprintln(color.Output, "Series popularity:", "#"+red(details.Popularity))
	fmt.Fprintln(color.Output, "Series rating:", "#"+yellow(details.Rating))
	fmt.Fprintln(color.Output, "Duration:", yellow(details.Duration))
	fmt.Fprintln(color.Output, "Genres:")
	printSlice(formatGenres(details.Genres, yellow))

	fmt.Fprintln(color.Output, "Episodes:", red(entry.WatchedEpisodes), "/", red(entry.Episodes))
	fmt.Fprintln(color.Output, "Score:", red(entry.MyScore))
	fmt.Fprintln(color.Output, "Status:", yellow(entry.MyStatus))
	fmt.Fprintln(color.Output, "Last updated:", red(time.Unix(entry.LastUpdated, 0)))
	fmt.Fprintln(color.Output, "Website url:", cyan(cfg.Websites[entry.ID]))

	fmt.Fprintln(color.Output)

	fmt.Fprintln(color.Output, "Synposis:", green(details.Synopsis))

	return nil
}

type sPrintFunc func(a ...interface{}) string

func formatSynonyms(synonyms string, sPrintFunc sPrintFunc) []string {
	synoSplit := strings.Split(synonyms, ";")
	for i, length := 0, len(synoSplit); i < length; {
		if synoSplit[i] == "" {
			synoSplit = synoSplit[:i+copy(synoSplit[i:], synoSplit[i+1:])]
			length--
		} else {
			synoSplit[i] = sPrintFunc(strings.TrimSpace(synoSplit[i]))
			i++
		}
	}
	return synoSplit
}

func formatGenres(genres []string, sPrintFunc sPrintFunc) []string {
	if length := len(genres); length == 0 {
		return genres
	} else if length == 1 {
		genres[0] = strings.Trim(genres[0], "[]")
	} else {
		genres[0] = strings.TrimLeft(genres[0], "[")
		genres[length-1] = strings.TrimRight(genres[length-1], "]")
	}
	for i := range genres {
		genres[i] = sPrintFunc(genres[i])
	}
	return genres
}

func printRelated(ctx *cli.Context) error {
	cfg := LoadConfig()
	c, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}

	selEntry := list.GetByID(cfg.SelectedID)
	if selEntry == nil {
		return fmt.Errorf("no entry selected")
	}

	details, err := mal.FetchDetailsWithAnimation(c, selEntry)
	if err != nil {
		return err
	}

	for _, related := range details.Related {
		title := color.HiYellowString("%s", related.Title)
		fmt.Fprintf(color.Output, "%s: %s (%s)\n", related.Relation, title, related.Url)
	}

	return nil
}

func printMusic(ctx *cli.Context) error {
	cfg := LoadConfig()
	c, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}

	entry := list.GetByID(cfg.SelectedID)
	if entry == nil {
		return fmt.Errorf("no entry selected")
	}

	details, err := mal.FetchDetailsWithAnimation(c, entry)

	printThemes := func(themes []string) {
		for _, theme := range themes {
			fmt.Fprintf(
				color.Output, "  %s\n",
				color.HiYellowString("%s", strings.TrimSpace(theme)))
		}
	}

	fmt.Fprintln(color.Output, "Openings:")
	printThemes(details.OpeningThemes)

	fmt.Fprintln(color.Output, "\nEndings:")
	printThemes(details.EndingThemes)

	return nil
}

func printBroadcast(ctx *cli.Context) error {
	cfg := LoadConfig()
	c, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}

	entry := list.GetByID(cfg.SelectedID)
	if entry == nil {
		return fmt.Errorf("no entry selected")
	}

	yellow := color.New(color.FgHiYellow).SprintFunc()
	green := color.New(color.FgHiGreen).SprintFunc()

	if entry.Status != mal.CurrentlyAiring {
		return fmt.Errorf("%s isn't currently airing", yellow(entry.Title))
	}

	details, err := mal.FetchDetailsWithAnimation(c, entry)
	if err != nil {
		return err
	}

	fmt.Fprintf(color.Output, "Title: %s\nBroadcast: %s\n",
		yellow(entry.Title),
		green(details.Broadcast))

	return nil
}

func copyIntoClipboard(ctx *cli.Context) error {
	cfg := LoadConfig()
	_, list, err := loadMAL(ctx)
	if err != nil {
		return err
	}

	entry := list.GetByID(cfg.SelectedID)
	if entry == nil {
		return fmt.Errorf("no entry selected")
	}

	var text string

	switch strings.ToLower(ctx.Args().First()) {
	case "title":
		text = entry.Title
	case "url":
		entryUrl, ok := cfg.Websites[cfg.SelectedID]
		if !ok {
			return fmt.Errorf("no url to copy")
		}
		text = entryUrl
	default:
		return fmt.Errorf("usage: mal copy [title|url]")
	}

	if err = clipboard.WriteAll(text); err == nil {
		fmt.Fprintln(color.Output, "Text", color.HiYellowString("%s", text), "copied into clipboard")
	}

	return err
}
