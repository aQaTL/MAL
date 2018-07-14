package main

import (
	"fmt"
	"github.com/aqatl/mal/anilist"
	"github.com/fatih/color"
	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

func AniListApp(app *cli.App) *cli.App {
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "max",
			Usage: "visible entries threshold",
		},
		cli.BoolFlag{
			Name:  "all, a",
			Usage: "display all entries; same as --max -1",
		},
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:      "mal",
			Aliases:   []string{"s"},
			Usage:     "Switches app mode to MyAnimeList",
			UsageText: "mal mal",
			Action:    switchToMal,
		},
		cli.Command{
			Name:     "eps",
			Aliases:  []string{"episodes"},
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
			Name:      "score",
			Category:  "Update",
			Usage:     "Set your rating for selected entry",
			UsageText: "mal score <0-10>",
			Action:    alSetEntryScore,
		},
		cli.Command{
			Name:      "sel",
			Aliases:   []string{"select"},
			Category:  "Config",
			Usage:     "Select an entry",
			UsageText: "mal sel [entry title]",
			Action:    alSelectEntry,
		},
		cli.Command{
			Name:     "nyaa",
			Aliases:  []string{"n"},
			Category: "Action",
			Usage:    "Open interactive torrent search",
			Action:   alNyaaCui,
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
			Name:      "airing",
			Aliases:   []string{"broadcast"},
			Category:  "Action",
			Usage:     "Print airing time of next episode",
			UsageText: "mal broadcast",
			Action:    alNextAiringEpisode,
		},
	}

	app.Action = cli.ActionFunc(aniListDefaultAction)

	return app
}

func aniListDefaultAction(ctx *cli.Context) error {
	al, err := loadAniList()
	if err != nil {
		return err
	}
	cfg := LoadConfig()
	list := alGetList(al, cfg.ALStatus)

	sort.Slice(list, func(i, j int) bool {
		return list[i].UpdatedAt > list[j].UpdatedAt
	})

	var visibleEntries int
	if visibleEntries = ctx.Int("max"); visibleEntries == 0 {
		//`Max` flag not specified, get value from config
		visibleEntries = cfg.MaxVisibleEntries
	}
	if visibleEntries > len(list) || visibleEntries < 0 || ctx.Bool("all") {
		visibleEntries = len(list)
	}

	fmt.Printf("No%64s%8s%6s\n", "Title", "Eps", "Score")
	fmt.Println("================================================================================")
	pattern := "%2d%64.64s%8s%6d\n"
	var entry *anilist.MediaListEntry
	for i := visibleEntries - 1; i >= 0; i-- {
		entry = &list[i]
		if entry.ListId == cfg.ALSelectedID {
			color.HiYellow(pattern, i+1, entry.Title.UserPreferred,
				fmt.Sprintf("%d/%d", entry.Progress, entry.Episodes),
				entry.Score)
		} else {
			fmt.Printf(pattern, i+1, entry.Title.UserPreferred,
				fmt.Sprintf("%d/%d", entry.Progress, entry.Episodes),
				entry.Score)
		}
	}

	return nil
}

func alGetList(al *AniList, status anilist.MediaListStatus) List {
	if status == anilist.All {
		return al.List
	} else {
		list := make(List, 0)
		for i := range al.List {
			if al.List[i].Status == status {
				list = append(list, al.List[i])
			}
		}
		return list
	}
}

func loadAniListFull() (al *AniList, entry *anilist.MediaListEntry, cfg *Config, err error) {
	al, err = loadAniList()
	if err != nil {
		return
	}
	cfg = LoadConfig()
	if cfg.ALSelectedID == 0 {
		fmt.Println("No entry selected")
	}
	entry = al.GetMediaListById(cfg.ALSelectedID)
	if entry == nil {
		err = fmt.Errorf("no entry found")
	}
	return
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
	al, entry, cfg, err := loadAniListFull()
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
	} else {
		entry.Progress++
	}

	alStatusAutoUpdate(cfg, entry)

	if err = anilist.SaveMediaListEntry(entry, al.Token); err != nil {
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
	al, entry, _, err := loadAniListFull()
	if err != nil {
		return err
	}

	status := anilist.ParseStatus(ctx.Args().First())
	if status == anilist.All {
		return fmt.Errorf("invalid status; possible values: " +
			"watching|planning|completed|dropped|paused|repeating")
	}

	entry.Status = status

	if err = anilist.SaveMediaListEntry(entry, al.Token); err != nil {
		return err
	}
	if err = saveAniListAnimeLists(al); err != nil {
		return err
	}

	fmt.Println("Updated successfully")
	alPrintEntryDetails(entry)
	return nil
}

func alSetEntryScore(ctx *cli.Context) error {
	al, entry, _, err := loadAniListFull()
	if err != nil {
		return err
	}

	score, err := strconv.Atoi(ctx.Args().First())
	if err != nil || score < 0 || score > 10 {
		return fmt.Errorf("invalid score; valid range: <0;10>")
	}

	entry.Score = score

	if err = anilist.SaveMediaListEntry(entry, al.Token); err != nil {
		return err
	}
	if err = saveAniListAnimeLists(al); err != nil {
		return err
	}

	fmt.Println("Updated successfully")
	alPrintEntryDetails(entry)
	return nil
}

func alSelectEntry(ctx *cli.Context) error {
	al, err := loadAniList()
	if err != nil {
		return err
	}
	cfg := LoadConfig()

	searchTerm := strings.ToLower(strings.Join(ctx.Args(), " "))
	if searchTerm == "" {
		return alFuzzySelectEntry(ctx)
	}

	for _, entry := range al.List {
		title := entry.Title.Romaji + " " + entry.Title.English + " " + entry.Title.Native
		if strings.Contains(strings.ToLower(title), searchTerm) {
			alSaveSelection(cfg, &entry)
			return nil
		}
	}

	return alFuzzySelectEntry(ctx)
}

func alSaveSelection(cfg *Config, entry *anilist.MediaListEntry) {
	cfg.ALSelectedID = entry.ListId
	cfg.Save()

	fmt.Println("Selected entry:")
	alPrintEntryDetails(entry)
}

func alNyaaWebsite(ctx *cli.Context) error {
	al, err := loadAniList()
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
		alts := make([]string, 0, 3+len(entry.Synonyms))
		if t := entry.Title.English; t != "" {
			alts = append(alts, t)
		}
		if t := entry.Title.Native; t != "" {
			alts = append(alts, t)
		}
		if t := entry.Title.Romaji; t != "" {
			alts = append(alts, t)
		}
		alts = append(alts, entry.Synonyms...)
		fmt.Printf("Select desired title\n\n")
		if searchTerm = chooseStrFromSlice(alts); searchTerm == "" {
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
	al, err := loadAniList()
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
		color.HiYellow("%v", entry.Title)
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
	al, err := loadAniList()
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

func alNextAiringEpisode(ctx *cli.Context) error {
	al, entry, cfg, err := loadAniListFull()
	if err != nil {
		return err
	}

	episode := entry.Progress
	if cfg.StatusAutoUpdateMode != AfterThreshold && entry.Progress < entry.Episodes {
		episode++
	}
	schedule, err := anilist.QueryAiringSchedule(entry.Id, episode, al.Token)
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
		cyan(airingAt.Format("15:04:05 02-01-2006 MST")),
	)

	tua := schedule.TimeUntilAiring
	if tua < 0 {
		tua *= -1
	}
	timeUntilAiring, err := time.ParseDuration(strconv.Itoa(tua) + "s")
	if err != nil {
		fmt.Println(err)
	} else if schedule.TimeUntilAiring < 0 {
		fmt.Fprintln(color.Output, "Episode aired", cyan(timeUntilAiring), "ago")
	} else {
		fmt.Fprintln(color.Output, "Time until airing:", cyan(timeUntilAiring))
	}
	return nil
}
