package main

import (
	"os"
	"fmt"
	"github.com/fatih/color"
	"path/filepath"
	"github.com/urfave/cli"
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
	MalMode     Mode = iota
	AniListMode
)

type AppConfig struct {
	Mode Mode
}

func main() {
	checkDataDir()

	appCfg := AppConfig{}
	LoadJsonFile(AppConfigFile, &appCfg)

	switch appCfg.Mode {
	case MalMode:
		runApp(MalApp())
	case AniListMode:
		runApp(AniListApp())
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
