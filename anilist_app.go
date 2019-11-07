package main

import (
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aqatl/mal/anilist"
	"github.com/aqatl/mal/mal"
	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/hako/durafmt"
	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli"
)

func AniListApp(app *cli.App) *cli.App {
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "r, refresh",
			Usage: "refreshes cached list",
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
				"[watching|planning|completed|repeating|paused|dropped]",
		},
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:      "switch",
			Aliases:   []string{"s"},
			Usage:     "Switches app mode between Anilist and MyAnimeList",
			UsageText: "mal mal",
			Action:    switchToMal,
		},
		cli.Command{
			Name:     "eps",
			Aliases:  []string{"episodes", "e"},
			Category: "Update",
			Usage: "Set the watched episodes value. " +
				"If n not specified, the number will be increased by one",
			UsageText: "mal eps <n>",
			Action:    alSetEntryEpisodes,
		},
		cli.Command{
			Name:      "status",
			Category:  "Update",
			Usage:     "Set your status for selected entry",
			UsageText: "mal status [watching|planning|completed|dropped|paused|repeating]",
			Action:    alSetEntryStatus,
		},
		cli.Command{
			Name:      "cmpl",
			Category:  "Update",
			Usage:     "Set entry status to completed",
			UsageText: "mal cmpl",
			Action:    alSetEntryStatusCompleted,
		},
		cli.Command{
			Name:      "score",
			Category:  "Update",
			Usage:     "Set your rating for selected entry",
			UsageText: "mal score <0-10>",
			Action:    alSetEntryScore,
		},
		cli.Command{
			Name:      "delete",
			Aliases:   []string{"del"},
			Category:  "Update",
			Usage:     "Delete entry",
			UsageText: "mal del",
			Action:    alDeleteEntry,
		},
		cli.Command{
			Name:      "sel",
			Aliases:   []string{"select"},
			Category:  "Config",
			Usage:     "Select an entry",
			UsageText: "mal sel [entry title]",
			Action:    alSelectEntry,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "rand",
					Usage: "select random entry from \"planning\" list",
				},
			},
		},
		cli.Command{
			Name:      "selected",
			Aliases:   []string{"curr"},
			Category:  "Action",
			Usage:     "Display info about currently selected entry",
			UsageText: "mal curr",
			Action:    alShowSelectedEntry,
		},
		cli.Command{
			Name:     "nyaa",
			Aliases:  []string{"n"},
			Category: "Action",
			Usage:    "Open interactive torrent search",
			Action:   alNyaaCui,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "alt",
					Usage: "choose an alternative title",
				},
				cli.StringFlag{
					Name:  "custom",
					Usage: "Adds custom nyaa search query for the selected entry",
				},
			},
		},
		cli.Command{
			Name:      "nyaa-web",
			Aliases:   []string{"nw"},
			Category:  "Action",
			Usage:     "Open torrent search in browser",
			UsageText: "mal nyaa-web",
			Action:    alNyaaWebsite,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "alt",
					Usage: "choose an alternative title",
				},
			},
		},
		cli.Command{
			Name:      "web",
			Aliases:   []string{"website", "open", "url"},
			Category:  "Action",
			Usage:     "Open url associated with selected entry or change url if provided",
			UsageText: "mal web <url>",
			Action:    alOpenWebsite,
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
					Action:    alPrintWebsites,
				},
			},
		},
		cli.Command{
			Name:            "search",
			Aliases:         []string{"add", "browse"},
			Category:        "Action",
			Usage:           "Browse online database and new entries",
			UsageText:       "mal search [search query]",
			Action:          alSearch,
			SkipFlagParsing: true,
		},
		cli.Command{
			Name:      "stats",
			Category:  "Action",
			Usage:     "Show your account statistics",
			UsageText: "mal stats",
			Action:    alStats,
		},
		cli.Command{
			Name:      "airing",
			Aliases:   []string{"broadcast"},
			Category:  "Action",
			Usage:     "Print airing time of next episode",
			UsageText: "mal airing [episode]",
			Action:    alAiringTime,
		},
		cli.Command{
			Name:      "music",
			Category:  "Action",
			Usage:     "Print opening and ending themes",
			UsageText: "mal music",
			Action:    alPrintMusic,
		},
		cli.Command{
			Name:      "copy",
			Category:  "Action",
			Usage:     "Copy selected value into system clipboard",
			UsageText: "mal copy [title|url]",
			Action:    alCopyIntoClipboard,
		},
		cli.Command{
			Name:      "anilist",
			Aliases:   []string{"al"},
			Category:  "Action",
			Usage:     "Open selected entry's AniList site",
			UsageText: "mal al",
			Action:    alOpenEntrySite,
		},
		cli.Command{
			Name:      "mal",
			Category:  "Action",
			Usage:     "Open selected entry's MyAnimeList site",
			UsageText: "mal mal",
			Action:    alOpenMalSite,
		},
		cli.Command{
			Name:      "airnot",
			Category:  "Action",
			Usage:     "Fetch airing notifications",
			UsageText: "mal airnot",
			Action:    alAiringNotifications,
			Flags: []cli.Flag{
				cli.UintFlag{
					Name:  "max",
					Usage: "Set max amount of notifications displayed",
					Value: 50,
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
					Name:            "list-width",
					Usage:           "Change the width of displayed list",
					UsageText:       "mal cfg list-width [width]",
					SkipFlagParsing: true,
					Action:          configChangeListWidth,
				},
				cli.Command{
					Name:      "status",
					Usage:     "Status value of displayed entries",
					UsageText: "mal cfg status [watching|planning|completed|dropped|paused|repeating]",
					Action:    configChangeAlStatus,
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
	}

	app.Action = cli.ActionFunc(aniListDefaultAction)

	return app
}

func aniListDefaultAction(ctx *cli.Context) error {
	al, err := loadAniList(ctx)
	if err != nil {
		return err
	}
	cfg := LoadConfig()
	status := cfg.ALStatus
	if statusFlag := ctx.String("status"); statusFlag != "" {
		status = anilist.ParseStatus(statusFlag)
	}
	list := alGetList(al, status)

	sort.Slice(list, func(i, j int) bool {
		return list[i].UpdatedAt > list[j].UpdatedAt
	})

	var visibleEntries int
	if visibleEntries = ctx.Int("max"); visibleEntries == 0 {
		// `Max` flag not specified, get value from config
		visibleEntries = cfg.MaxVisibleEntries
	}
	if visibleEntries > len(list) || visibleEntries < 0 || ctx.Bool("all") {
		visibleEntries = len(list)
	}

	numberFieldWidth := int(math.Max(math.Ceil(math.Log10(float64(visibleEntries+1))), 2))
	titleWidth := cfg.ListWidth - numberFieldWidth - 8 - 6
	fmt.Printf("%*s%*.*s%8s%6s\n",
		numberFieldWidth, "No", titleWidth, titleWidth, "Title", "Eps", "Score")
	fmt.Println(strings.Repeat("=", cfg.ListWidth))
	pattern := "%*d%*.*s%8s%6d\n"
	var entry *anilist.MediaListEntry
	for i := visibleEntries - 1; i >= 0; i-- {
		entry = &list[i]
		if entry.Id == cfg.ALSelectedID {
			color.HiYellow(pattern, numberFieldWidth, i+1, titleWidth, titleWidth,
				entry.Title.UserPreferred,
				fmt.Sprintf("%d/%d", entry.Progress, entry.Episodes),
				entry.Score)
		} else {
			fmt.Printf(pattern, numberFieldWidth, i+1, titleWidth, titleWidth,
				entry.Title.UserPreferred,
				fmt.Sprintf("%d/%d", entry.Progress, entry.Episodes),
				entry.Score)
		}
	}

	return nil
}

func switchToMal(ctx *cli.Context) error {
	appCfg := AppConfig{}
	LoadJsonFile(AppConfigFile, &appCfg)
	appCfg.Mode = MalMode
	if err := SaveJsonFile(AppConfigFile, &appCfg); err != nil {
		return err
	}
	fmt.Println("App mode switched to MyAnimeList")
	return nil
}

func alSetEntryEpisodes(ctx *cli.Context) error {
	al, entry, cfg, err := loadAniListFull(ctx)
	if err != nil {
		return err
	}
	epsBefore := entry.Progress

	if arg := ctx.Args().First(); arg != "" {
		n, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("n must be a non-negative integer")
		}
		if n < 0 {
			return fmt.Errorf("n can't be lower than 0")
		}
		entry.Progress = n
	} else if cfg.StatusAutoUpdateMode == AfterThreshold && entry.Progress == 0 {
		entry.Progress += 2
	} else {
		entry.Progress++
	}

	alStatusAutoUpdate(cfg, entry)

	if err = anilist.SaveMediaListEntryWaitAnimation(entry, al.Token); err != nil {
		return err
	}
	if err = saveAniListAnimeLists(al); err != nil {
		return err
	}

	fmt.Println("Updated successfully")
	alPrintEntryDetailsAfterUpdatedEpisodes(entry, epsBefore)
	return nil
}

func alStatusAutoUpdate(cfg *Config, entry *anilist.MediaListEntry) {
	if cfg.StatusAutoUpdateMode == Off || entry.Episodes == 0 {
		return
	}

	if (cfg.StatusAutoUpdateMode == Normal && entry.Progress >= entry.Episodes) ||
		(cfg.StatusAutoUpdateMode == AfterThreshold && entry.Progress > entry.Episodes) {
		entry.Status = anilist.Completed
		entry.Progress = entry.Episodes
		return
	}

	if entry.Status == anilist.Completed && entry.Progress < entry.Episodes {
		entry.Status = anilist.Current
		return
	}
}

func alSetEntryStatus(ctx *cli.Context) error {
	al, entry, _, err := loadAniListFull(ctx)
	if err != nil {
		return err
	}

	status := anilist.ParseStatus(ctx.Args().First())
	if status == anilist.All {
		return fmt.Errorf("invalid status; possible values: " +
			"watching|planning|completed|dropped|paused|repeating")
	}

	entry.Status = status

	if err = anilist.SaveMediaListEntryWaitAnimation(entry, al.Token); err != nil {
		return err
	}
	if err = saveAniListAnimeLists(al); err != nil {
		return err
	}

	fmt.Println("Updated successfully")
	alPrintEntryDetails(entry)
	return nil
}

func alSetEntryStatusCompleted(ctx *cli.Context) error {
	return ctx.App.Run([]string{"", "status", "completed"})
}

func alSetEntryScore(ctx *cli.Context) error {
	al, entry, _, err := loadAniListFull(ctx)
	if err != nil {
		return err
	}

	score, err := strconv.Atoi(ctx.Args().First())
	if err != nil || score < 0 || score > 10 {
		return fmt.Errorf("invalid score; valid range: <0;10>")
	}

	entry.Score = score

	if err = anilist.SaveMediaListEntryWaitAnimation(entry, al.Token); err != nil {
		return err
	}
	if err = saveAniListAnimeLists(al); err != nil {
		return err
	}

	fmt.Println("Updated successfully")
	alPrintEntryDetails(entry)
	return nil
}

func alDeleteEntry(ctx *cli.Context) error {
	al, entry, _, err := loadAniListFull(ctx)
	if err != nil {
		return err
	}

	if err := anilist.DeleteMediaListEntry(entry, al.Token); err != nil {
		return err
	}

	fmt.Println("Entry deleted successfully")
	alPrintEntryDetails(entry)

	al.List = al.List.DeleteById(entry.ListId)
	return saveAniListAnimeLists(al)
}

func alSelectEntry(ctx *cli.Context) error {
	if ctx.Bool("rand") {
		return alSelectRandomEntry(ctx)
	}

	al, err := loadAniList(ctx)
	if err != nil {
		return err
	}
	cfg := LoadConfig()

	searchTerm := strings.ToLower(strings.Join(ctx.Args(), " "))
	if searchTerm == "" {
		return alFuzzySelectEntry(ctx)
	}

	var matchedEntry *anilist.MediaListEntry = nil
	for i, entry := range al.List {
		title := entry.Title.Romaji + " " + entry.Title.English + " " + entry.Title.Native
		if strings.Contains(strings.ToLower(title), searchTerm) {
			if matchedEntry != nil {
				matchedEntry = nil
				break
			}
			matchedEntry = &al.List[i]
		}
	}
	if matchedEntry != nil {
		alSaveSelection(cfg, matchedEntry)
		return nil
	}

	return alFuzzySelectEntry(ctx)
}

func alSelectRandomEntry(ctx *cli.Context) error {
	al, err := loadAniList(ctx)
	if err != nil {
		return err
	}

	planToWatchList := alGetList(al, anilist.Planning)
	idx := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(planToWatchList))
	alSaveSelection(LoadConfig(), &planToWatchList[idx])

	return nil
}

func alSaveSelection(cfg *Config, entry *anilist.MediaListEntry) {
	cfg.ALSelectedID = entry.Id
	cfg.Save()

	fmt.Println("Selected entry:")
	alPrintEntryDetails(entry)
}

func alShowSelectedEntry(ctx *cli.Context) error {
	_, entry, _, err := loadAniListFull(ctx)
	if err != nil {
		return err
	}
	alPrintEntryDetails(entry)
	return nil
}

func alNyaaWebsite(ctx *cli.Context) error {
	al, err := loadAniList(ctx)
	if err != nil {
		return err
	}
	cfg := LoadConfig()

	entry := al.GetMediaListById(cfg.ALSelectedID)
	if entry == nil {
		return fmt.Errorf("no entry selected")
	}

	var searchTerm string
	if ctx.Bool("alt") {
		fmt.Printf("Select desired title\n\n")
		if searchTerm = chooseStrFromSlice(sliceOfEntryTitles(entry)); searchTerm == "" {
			return fmt.Errorf("no alternative titles")
		}
	} else {
		searchTerm = entry.Title.Romaji
	}

	address := "https://nyaa.si/?f=0&c=1_2&q=" + url.QueryEscape(searchTerm)
	if path := cfg.BrowserPath; path == "" {
		open.Start(address)
	} else {
		open.StartWith(address, path)
	}

	fmt.Println("Searched for:")
	alPrintEntryDetails(entry)
	return nil
}

func alOpenWebsite(ctx *cli.Context) error {
	al, err := loadAniList(ctx)
	if err != nil {
		return nil
	}

	cfg := LoadConfig()

	entry := al.GetMediaListById(cfg.ALSelectedID)
	if entry == nil {
		return fmt.Errorf("no entry selected")
	}

	if newUrl := ctx.Args().First(); newUrl != "" {
		cfg.Websites[entry.IdMal] = newUrl
		cfg.Save()

		fmt.Print("Entry: ")
		color.HiYellow("%s", entry.Title.UserPreferred)
		fmt.Print("URL: ")
		color.HiRed("%v", cfg.Websites[entry.IdMal])

		return nil
	}

	if ctx.Bool("clear") {
		delete(cfg.Websites, entry.IdMal)
		cfg.Save()

		fmt.Println("Entry cleared")
		return nil
	}

	if entryUrl, ok := cfg.Websites[entry.IdMal]; ok {
		if path := cfg.BrowserPath; path == "" {
			open.Start(entryUrl)
		} else {
			open.StartWith(entryUrl, path)
		}

		fmt.Println("Opened website for:")
		alPrintEntryDetails(entry)
		fmt.Fprintf(color.Output, "URL: %v\n", color.CyanString("%v", entryUrl))
	} else {
		fmt.Println("Nothing to open")
	}

	return nil
}

func alPrintWebsites(ctx *cli.Context) error {
	al, err := loadAniList(ctx)
	if err != nil {
		return err
	}

	cfg := LoadConfig()

	for k, v := range cfg.Websites {
		entryUrl := fmt.Sprintf("\033[3%d;%dm%s\033[0m ", 3, 1, v)

		var title string
		if entry := al.GetMediaListByMalId(k); entry != nil {
			title = entry.Title.UserPreferred
		}

		fmt.Fprintf(color.Output, "%6d (%s): %s\n", k, title, entryUrl)
	}

	return nil
}

func alStats(ctx *cli.Context) error {
	al, err := loadAniList(ctx)
	if err != nil {
		return err
	}

	yellow := color.New(color.FgHiYellow).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()
	cyan := color.New(color.FgHiCyan).SprintFunc()
	magenta := color.New(color.FgHiMagenta).SprintFunc()

	lists := [6]List{
		alGetList(al, anilist.Current),
		alGetList(al, anilist.Planning),
		alGetList(al, anilist.Completed),
		alGetList(al, anilist.Repeating),
		alGetList(al, anilist.Paused),
		alGetList(al, anilist.Dropped),
	}

	totalShows := 0
	totalTimeSpentWatching := 0
	totalEpisodesWatched := 0
	for _, list := range lists {
		if len(list) == 0 {
			continue
		}
		totalShows += len(list)
		episodesWatched := 0
		timeSpentWatching := 0
		for _, entry := range list {
			timeSpentWatching += (entry.Progress * entry.Duration) +
				(entry.Repeat * entry.Episodes * entry.Duration)
			episodesWatched += entry.Progress + entry.Episodes*entry.Repeat
		}
		totalTimeSpentWatching += timeSpentWatching
		totalEpisodesWatched += episodesWatched

		timeSpentWatchingFormatted, _ := durafmt.ParseString(fmt.Sprint(timeSpentWatching, "m"))

		fmt.Fprintf(color.Output,
			`%s:
  entries: %s
  episodes: %s
  time spent watching: %s
`,
			list[0].Status.String(),
			red(len(list)),
			magenta(episodesWatched),
			cyan(timeSpentWatchingFormatted),
		)
	}

	totalTimeSpentWatchingDuration, _ := time.ParseDuration(
		fmt.Sprint(totalTimeSpentWatching, "m"))

	fmt.Println()
	fmt.Fprintln(color.Output, "Total episodes watched:", red(totalEpisodesWatched))
	fmt.Fprintln(color.Output, "Total shows:", red(totalShows))
	fmt.Fprintf(color.Output,
		"Total time spent watching: %s (%s days)\n",
		yellow(durafmt.Parse(totalTimeSpentWatchingDuration).String()),
		cyan(int(totalTimeSpentWatchingDuration.Hours()/24+0.5)))

	return nil
}

func alAiringTime(ctx *cli.Context) error {
	al, entry, cfg, err := loadAniListFull(ctx)
	if err != nil {
		return err
	}

	var episode int
	if argsLen := ctx.NArg(); argsLen == 1 {
		episode, err = strconv.Atoi(ctx.Args().First())
		if err != nil {
			return err
		}
	} else if argsLen > 1 {
		return fmt.Errorf("too many arguments")
	} else {
		episode = entry.Progress
		if episode == 0 ||
			cfg.StatusAutoUpdateMode != AfterThreshold && entry.Progress < entry.Episodes {

			episode++
		}
	}

	schedule, err := anilist.QueryAiringScheduleWaitAnimation(entry.Id, episode, al.Token)
	if err != nil {
		return err
	}

	airingAt := time.Unix(int64(schedule.AiringAt), 0)

	yellow := color.New(color.FgHiYellow).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()
	cyan := color.New(color.FgHiCyan).SprintFunc()
	fmt.Fprintf(
		color.Output,
		"Title: %s\n"+
			"Episode: %s\n"+
			"Airing at: %s\n",
		yellow(entry.Title.UserPreferred),
		red(schedule.Episode),
		cyan(airingAt.Format("15:04:05 02-01-2006 MST (Monday)")),
	)

	tua := schedule.TimeUntilAiring
	if tua < 0 {
		tua *= -1
	}
	timeUntilAiring, err := durafmt.ParseString(strconv.Itoa(tua) + "s")
	if err != nil {
		fmt.Println(err)
	} else if schedule.TimeUntilAiring < 0 {
		fmt.Fprintln(color.Output, "Episode aired", cyan(timeUntilAiring), "ago")
	} else {
		fmt.Fprintln(color.Output, "Time until airing:", cyan(timeUntilAiring))
	}
	return nil
}

func alPrintMusic(ctx *cli.Context) error {
	_, entry, _, err := loadAniListFull(ctx)
	if err != nil {
		return err
	}

	details, err := mal.FetchDetailsWithAnimation(&mal.Client{}, &mal.Anime{ID: entry.IdMal})
	if err != nil {
		return err
	}

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

func alCopyIntoClipboard(ctx *cli.Context) error {
	_, entry, cfg, err := loadAniListFull(ctx)
	if err != nil {
		return err
	}

	var text string

	switch strings.ToLower(ctx.Args().First()) {
	case "title":
		alts := sliceOfEntryTitles(entry)
		fmt.Printf("Select desired title\n\n")
		if text = chooseStrFromSlice(alts); text == "" {
			return fmt.Errorf("no alternative titles")
		}
	case "url":
		entryUrl, ok := cfg.Websites[entry.IdMal]
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

func alOpenEntrySite(ctx *cli.Context) error {
	_, entry, cfg, err := loadAniListFull(ctx)
	if err != nil {
		return err
	}

	uri := fmt.Sprintf("%s/%s/%d", anilist.ALDomain, strings.ToLower(entry.Type), entry.Id)
	if path := cfg.BrowserPath; path == "" {
		open.Start(uri)
	} else {
		open.StartWith(uri, path)
	}
	fmt.Println("Opened website for:")
	alPrintEntryDetails(entry)

	return nil
}

func alOpenMalSite(ctx *cli.Context) error {
	_, entry, cfg, err := loadAniListFull(ctx)
	if err != nil {
		return err
	}

	openMalSite(cfg, entry.IdMal)
	fmt.Println("Opened website for:")
	alPrintEntryDetails(entry)

	return nil
}

func alAiringNotifications(ctx *cli.Context) error {
	al, err := loadAniList(ctx)
	if err != nil {
		return err
	}

	notifications, err := anilist.QueryAiringNotificationsWaitAnimation(
		1, int(ctx.Uint("max")), false, al.Token)
	if err != nil {
		return err
	}
	sort.SliceStable(notifications, func(i, j int) bool {
		return notifications[i].CreatedAt < notifications[j].CreatedAt
	})

	cyan := color.New(color.FgHiCyan).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()
	yellow := color.New(color.FgHiYellow).SprintFunc()
	for _, n := range notifications {
		t := time.Unix(int64(n.CreatedAt), 0).Format("02-01-2006 15:04")
		fmt.Fprintf(color.Output, "[%s] Episode %s of %s aired\n",
			cyan(t), red(n.Episode), yellow(n.Title.UserPreferred))
	}

	return nil
}
