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
