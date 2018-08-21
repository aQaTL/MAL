package main

import (
	"os"
	"fmt"
	"github.com/fatih/color"
	"path/filepath"
	"github.com/urfave/cli"
	"log"
)

var dataDir = filepath.Join(homeDir(), ".mal")
var (
	AppConfigFile = filepath.Join(dataDir, "appConfig.json")

	MalCredentialsFile = filepath.Join(dataDir, "malCred.dat")
	MalCacheFile       = filepath.Join(dataDir, "malCache.xml")
	MalStatsCacheFile  = filepath.Join(dataDir, "malStats.xml")
	MalConfigFile      = filepath.Join(dataDir, "malConfig.json")

	AniListCredsFile = filepath.Join(dataDir, "aniListCreds.json")
	AniListUserFile  = filepath.Join(dataDir, "aniListUser.json")
	AniListCacheFile = filepath.Join(dataDir, "aniListCache.json")
)

type Mode uint

const (
	MalMode Mode = iota
	AniListMode
)

type AppConfig struct {
	Mode Mode
}

func main() {
	checkDataDir()

	appCfg := AppConfig{}
	LoadJsonFile(AppConfigFile, &appCfg)

	app := cli.NewApp()
	app.Name = "mal"
	app.Usage = "App for managing your MAL"
	app.Version = "4ever in beta"

	switch appCfg.Mode {
	case MalMode:
		runApp(MalApp(app))
	case AniListMode:
		runApp(AniListApp(app))
	}
}

func checkDataDir() {
	if err := os.Mkdir(dataDir, os.ModePerm); err == nil {
		log.Printf("Created cache directory at %s", dataDir)
	} else if !os.IsExist(err) {
		log.Printf("Error creating cache directory (%s): %v", dataDir, err)
	}
}

func runApp(app *cli.App) {
	if err := app.Run(os.Args); err != nil {
		exitWithError(err)
	}
}

func exitWithError(err error) {
	fmt.Fprintf(color.Output, "Error: %v\n", err)
	os.Exit(1)
}
