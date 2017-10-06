package main

import (
	"github.com/aqatl/mal/mal"
	"github.com/urfave/cli"
	"os"
	"fmt"
	"log"
	"encoding/base64"
	"io/ioutil"
)

const CredentialsFile = "cred.dat"

func main() {
	app := cli.NewApp()
	app.Name = "mal"
	app.Usage = "App for managing your MAL"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "u, username",
			Usage: "specify username",
		},
		cli.StringFlag{
			Name: "p, password",
			Usage: "specify password",
		},
		cli.BoolFlag{
			Name: "c, cache",
			Usage: "Cache / save your password. Use with caution, your password can be decoded",
		},
	}

	app.Action = func(ctx *cli.Context) {
		credentials, err := ioutil.ReadFile(CredentialsFile)
		if err != nil {
			//credentials not found, using given username and password
			credentials = []byte(basicAuth(ctx.String("username"), ctx.String("password")))
		}
		c := mal.NewClient(string(credentials))

		list := c.AnimeList(mal.Completed)
		for _, anime := range list {
			fmt.Printf("%v\n", anime.Title)
		}

		if ctx.Bool("cache") {
			cacheCredentials(ctx.String("username"), ctx.String("password"))
		}
	}

	if err := app.Run(os.Args); err != nil {
		log.Printf("Arguments error: %v", err)
		os.Exit(1)
	}

}

func basicAuth(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func cacheCredentials(username, password string) {
	credentials := basicAuth(username, password)
	err := ioutil.WriteFile(CredentialsFile, []byte(credentials), 400)
	if err != nil {
		log.Printf("Caching credentials failed: %v", err)
	}
}
