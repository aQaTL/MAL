package main

import (
	"github.com/urfave/cli"
	"github.com/aqatl/mal/anilist"
	"fmt"
	"github.com/fatih/color"
	"sort"
	"net/url"
	"github.com/skratchdot/open-golang/open"
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
			Name:      "fuzzy-select",
			Aliases:   []string{"fsel"},
			Category:  "Config",
			Usage:     "Interactive fuzzy search through your list",
			UsageText: "mal fsel [search string (optional)]",
			Action:    alFuzzySelectEntry,
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

	sort.Slice(list.Entries, func(i, j int) bool {
		return list.Entries[i].UpdatedAt > list.Entries[j].UpdatedAt
	})

	var visibleEntries int
	if visibleEntries = ctx.Int("max"); visibleEntries == 0 {
		//`Max` flag not specified, get value from config
		visibleEntries = cfg.MaxVisibleEntries
	}
	if visibleEntries > len(list.Entries) || visibleEntries < 0 || ctx.Bool("all") {
		visibleEntries = len(list.Entries)
	}

	fmt.Printf("No%64s%8s%6s\n", "Title", "Eps", "Score")
	fmt.Println("================================================================================")
	pattern := "%2d%64.64s%8s%6d\n"
	var entry *anilist.MediaList
	for i := visibleEntries - 1; i >= 0; i-- {
		entry = &list.Entries[i]
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

func alGetList(al *anilist.AniList, status anilist.MediaListStatus) anilist.MediaListGroup {
	var list anilist.MediaListGroup
	if status == anilist.All {
		for i := range al.Lists {
			list.Entries = append(list.Entries, al.Lists[i].Entries...)
		}
	} else {
		for i := range al.Lists {
			if al.Lists[i].Status == status {
				return al.Lists[i]
			}
		}
	}
	return list
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
		alts := make([]string, 0, 3 + len(entry.Synonyms))
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

	entry := al.GetMediaListByMalId(cfg.SelectedID)
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
