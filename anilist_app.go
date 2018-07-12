package main

import (
	"github.com/urfave/cli"
	"github.com/aqatl/mal/anilist"
	"fmt"
	"github.com/fatih/color"
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

	var list anilist.MediaListGroup
	for i := range al.Lists {
		if al.Lists[i].Status == cfg.ALStatus {
			list = al.Lists[i]
		}
	}

	var visibleEntries int
	if visibleEntries = ctx.Int("max"); visibleEntries == 0 {
		//`Max` flag not specified, get value from config
		visibleEntries = cfg.MaxVisibleEntries
	}
	if visibleEntries > len(list.Entries) || visibleEntries < 0 || ctx.Bool("all") {
		visibleEntries = len(list.Entries)
	}
	visibleList := list.Entries[:visibleEntries]
	last := len(visibleList) - 1
	for i := 0; i < len(visibleList)/2; i++ {
		visibleList[i], visibleList[last-i] = visibleList[last-i], visibleList[i]
	}

	fmt.Printf("No%64s%8s%6s\n", "Title", "Eps", "Score")
	fmt.Println("================================================================================")
	i := visibleEntries
	pattern := "%2d%64.64s%8s%6d\n"
	for _, entry := range visibleList {
		if entry.IdMal == cfg.SelectedID {
			color.HiYellow(pattern, i, entry.Title.UserPreferred,
				fmt.Sprintf("%d/%d", entry.Progress, entry.Episodes),
				entry.Score)
		} else {
			fmt.Printf(pattern, i, entry.Title.UserPreferred,
				fmt.Sprintf("%d/%d", entry.Progress, entry.Episodes),
				entry.Score)
		}
		i--
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
