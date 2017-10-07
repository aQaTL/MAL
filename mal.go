package main

import (
	"github.com/aqatl/mal/mal"
	"github.com/urfave/cli"
	"os"
	"log"
	"encoding/base64"
	"io/ioutil"
	"sort"
	"encoding/xml"
	"io"
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
		cli.BoolFlag{
			Name:  "no-verify",
			Usage: "don't verify credentials",
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

	var list []*mal.Anime

	if ctx.Bool("refresh") {
		list = c.AnimeList(mal.Watching)
	} else {
		list = loadCachedList()
	}

	sort.Sort(mal.AnimeSortByLastUpdated(list))
	list = list[:10]
	reverseAnimeSlice(list)

	PrettyList.Execute(os.Stdout, list)

	if ctx.Bool("save-password") {
		cacheCredentials(ctx.String("username"), ctx.String("password"))
	}

	if ctx.BoolT("cache") {
		cacheList(list)
	}
}

func basicAuth(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

func loadCachedList() []*mal.Anime {
	f, err := os.Open(MalCacheFile)
	defer f.Close()
	if err != nil {
		log.Printf("Error opening %s file: %v", MalCacheFile, err)
	}

	list := make([]*mal.Anime, 0)

	decoder := xml.NewDecoder(f)
	for t, err := decoder.Token(); err != io.EOF; t, err = decoder.Token() {
		if t, ok := t.(xml.StartElement); ok && t.Name.Local == "Anime" {
			var anime mal.Anime
			decoder.DecodeElement(&anime, &t)
			list = append(list, &anime)
		}
	}

	return list
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

func cacheList(list []*mal.Anime) {
	f, err := os.Create(MalCacheFile)
	defer f.Close()
	if err != nil {
		log.Printf("Error creating %s file: %v", MalCacheFile, err)
	}

	encoder := xml.NewEncoder(f)
	if err := encoder.Encode(list); err != nil {
		log.Printf("Caching error: %v", err)
	}
}
