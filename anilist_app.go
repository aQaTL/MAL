package main

import (
	"github.com/urfave/cli"
	"fmt"
)

func AniListApp(app *cli.App) *cli.App {
	app.Flags = []cli.Flag{

	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:      "mal",
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

	counter := 0
	for _, listGroup := range al.Lists {
		for _, entry := range listGroup.Entries {
			fmt.Println(entry.Title.Romaji)
			counter++
		}
	}

	fmt.Println(counter)

	return nil
}

func switchToMal(ctx *cli.Context) error {
	appCfg := AppConfig{}
	LoadJsonFile(AppConfigFile, &appCfg)
	appCfg.Mode = MalMode
	return SaveJsonFile(AppConfigFile, &appCfg)
}
