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
	"strconv"
)

const cacheDir = "data" + string(os.PathSeparator)
const CredentialsFile = cacheDir + "cred.dat"
const MalCacheFile = cacheDir + "cache.xml"
const ConfigFile = cacheDir + "config.json"

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
		cli.BoolFlag{
			Name:  "r, refresh",
			Usage: "refreshes cache file",
		},
		cli.BoolFlag{
			Name:  "no-verify",
			Usage: "don't verify credentials",
		},
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:     "inc",
			Aliases:  []string{"+1"},
			Category: "Update",
			Usage:    "Increment selected entry by one",
			UsageText: "mal inc",
			Action:   incrementEntry,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "n",
					Usage: "Specify exact episode to set the entry to",
				},
			},
		},
		cli.Command{
			Name:     "sel",
			Aliases:  []string{"select"},
			Category: "Config",
			Usage:    "Select an entry",
			UsageText: "mal sel [entry ID]",
			Action:   selectEntry,
		},
	}

	app.Action = defaultAction

	if err := app.Run(os.Args); err != nil {
		log.Printf("Arguments error: %v", err)
		os.Exit(1)
	}
}

func defaultAction(ctx *cli.Context) {
	c := mal.NewClient(credentials(ctx))

	config := LoadConfig()

	var list []*mal.Anime
	if ctx.Bool("refresh") || cacheNotExist() {
		list = c.AnimeList(mal.All)
	} else {
		list = loadCachedList()
	}
	sort.Sort(mal.AnimeSortByLastUpdated(list))

	visibleEntries := config.MaxVisibleEntries
	if visibleEntries > len(list) {
		visibleEntries = len(list)
	}
	visibleList := list[:visibleEntries]
	reverseAnimeSlice(visibleList)

	PrettyList.Execute(os.Stdout, PrettyListData{visibleList, config.SelectedID})

	if ctx.Bool("save-password") {
		cacheCredentials(ctx.String("username"), ctx.String("password"))
	}

	cacheList(list)
}

func credentials(ctx *cli.Context) string {
	credentials, err := ioutil.ReadFile(CredentialsFile)
	if err != nil {
		//credentials not found, using given username and password
		credentials = []byte(basicAuth(ctx.String("username"), ctx.String("password")))
	}
	return string(credentials)
}

func basicAuth(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

func cacheNotExist() bool {
	f, err := os.Open(MalCacheFile)
	defer f.Close()
	return os.IsNotExist(err)
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

func incrementEntry(ctx *cli.Context) error {
	c := mal.NewClient(credentials(ctx))
	cfg := LoadConfig()

	if cfg.SelectedID == 0 {
		log.Fatalln("No entry selected")
	}

	list := loadCachedList()
	var selectedEntry *mal.Anime
	for _, entry := range list {
		if entry.ID == cfg.SelectedID {
			selectedEntry = entry
			break
		}
	}

	if selectedEntry == nil {
		log.Fatalln("No entry found")
	}

	if ctx.Int("n") > 0 {
		selectedEntry.WatchedEpisodes = ctx.Int("n")
	} else {
		selectedEntry.WatchedEpisodes++
	}

	if c.Update(selectedEntry) {
		log.Printf("Updated successfully")
	}
	return nil
}

func selectEntry(ctx *cli.Context) error {
	cfg := LoadConfig()

	id, err := strconv.Atoi(ctx.Args().First())
	if err != nil {
		log.Fatalf("Error parsing id: %v", err)
	}

	cfg.SelectedID = id
	cfg.Save()
	return nil
}
