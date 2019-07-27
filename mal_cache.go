package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"github.com/aqatl/cliwait"
	"github.com/aqatl/mal/mal"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"
	"time"
)

func loadCredentials(ctx *cli.Context) string {
	if ctx.GlobalBool("prompt-credentials") { //Read credentials from console
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter username: ")
		username, _ := reader.ReadString('\n')

		fmt.Print("Enter password (chars hidden): ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			log.Printf("Error reading password: %v", err)
		}
		password := string(bytePassword)

		return basicAuth(strings.TrimSpace(username), strings.TrimSpace(password))
	} else { //Read credentials from CredentialsFile
		credentials, err := ioutil.ReadFile(MalCredentialsFile)
		if err != nil {
			log.Printf("Failed to load credentials: %v", err)
			return ""
		}
		return string(credentials)
	}
}

func saveCredentials(credentials string) {
	err := ioutil.WriteFile(MalCredentialsFile, []byte(credentials), 400)
	if err != nil {
		log.Printf("Caching credentials failed: %v", err)
	}
}

//Loads Client statistic data and returns Client's AnimeList
func loadData(c *mal.Client, ctx *cli.Context) (mal.AnimeList, error) {
	var list []*mal.Anime

	if ctx.GlobalBool("refresh") || cacheNotExist() {
		{
			var err error
			cliwait.DoFuncWithWaitAnimation("Fetching your list", func() {
				list, err = c.AnimeList(mal.All)
			})
			return nil, fmt.Errorf("error fetching your list\n%v", err)
		}

		cacheList(list)
		cacheClient(c)

		cfg := LoadConfig()
		cfg.LastUpdate = time.Now()
		cfg.Save()
	} else {
		list = loadCachedList()
		loadCachedStats(c)
	}
	return list, nil
}

func cacheNotExist() bool {
	f, err := os.Open(MalCacheFile)
	defer f.Close()
	f2, err2 := os.Open(MalStatsCacheFile)
	defer f2.Close()
	return os.IsNotExist(err) || os.IsNotExist(err2)
}

func cacheList(list []*mal.Anime) {
	f, err := os.Create(MalCacheFile)
	defer f.Close()
	if err != nil {
		log.Printf("Error creating %s file: %v", MalCacheFile, err)
		return
	}

	encoder := xml.NewEncoder(f)
	if err := encoder.Encode(list); err != nil {
		log.Printf("Encoding error: %v", err)
	}
}

func loadCachedList() mal.AnimeList {
	f, err := os.Open(MalCacheFile)
	defer f.Close()
	if err != nil {
		log.Printf("Error opening %s file: %v", MalCacheFile, err)
		return nil
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

func cacheClient(c *mal.Client) {
	f, err := os.Create(MalStatsCacheFile)
	defer f.Close()
	if err != nil {
		log.Printf("Error opening %s file: %v", MalStatsCacheFile, err)
		return
	}

	encoder := xml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		log.Printf("Encoding error: %v", err)
	}
}

func loadCachedStats(c *mal.Client) {
	f, err := os.Open(MalStatsCacheFile)
	defer f.Close()
	if err != nil {
		log.Printf("Error opening %s file: %v", MalCacheFile, err)
		return
	}

	decoder := xml.NewDecoder(f)
	if err := decoder.Decode(c); err != nil {
		log.Printf("Error decoding %s file", MalStatsCacheFile)
	}
}
