package main

import (
	"github.com/aqatl/mal/mal"
	"github.com/urfave/cli"
)

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
