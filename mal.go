package main

import (
	"github.com/aqatl/mal/mal"
	"github.com/urfave/cli"
	"os"
	"log"
	"encoding/base64"
	"io/ioutil"
	"sort"
)

const CredentialsFile = "cred.dat"
const MalCacheFile = "cache.json"

func main() {
	app := cli.NewApp()
	app.Name = "mal"
	app.Usage = "App for managing your MAL"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "u, username",
			Usage: "specify username",
		},
		cli.StringFlag{
			Name:  "p, password",
			Usage: "specify password",
		},
		cli.BoolFlag{
			Name:  "sp, save-password",
			Usage: "save your password. Use with caution, your password can be decoded",
		},
		cli.BoolTFlag{
			Name:  "cache",
			Usage: "cache your list",
		},
		cli.BoolFlag{
			Name:  "r, refresh",
			Usage: "refreshes cache file",
		},
	}

	app.Action = defaultAction

	if err := app.Run(os.Args); err != nil {
		log.Printf("Arguments error: %v", err)
		os.Exit(1)
	}

}

func defaultAction(ctx *cli.Context) {
	credentials, err := ioutil.ReadFile(CredentialsFile)
	if err != nil {
		//credentials not found, using given username and password
		credentials = []byte(basicAuth(ctx.String("username"), ctx.String("password")))
	}
	c := mal.NewClient(string(credentials))

	list := c.AnimeList(mal.Watching)
	sort.Sort(mal.AnimeSortByLastUpdated(list))
	list = list[:10]
	reverseAnimeSlice(list)

	PrettyList.Execute(os.Stdout, list)

	if ctx.Bool("save-password") {
		cacheCredentials(ctx.String("username"), ctx.String("password"))
	}

}

func basicAuth(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

func cacheCredentials(username, password string) {
	credentials := basicAuth(username, password)
	err := ioutil.WriteFile(CredentialsFile, []byte(credentials), 400)
	if err != nil {
		log.Printf("Caching credentials failed: %v", err)
	}
}

func reverseAnimeSlice(s []*mal.Anime) {
	last := len(s) - 1
	for i := 0; i < len(s)/2; i++ {
		s[i], s[last-i] = s[last-i], s[i]
	}
}
