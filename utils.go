package main

import (
	"os"
	"github.com/aqatl/mal/mal"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"github.com/urfave/cli"
	"encoding/base64"
	"os/user"
)

func credentials(ctx *cli.Context) string {
	credentials, err := ioutil.ReadFile(CredentialsFile)
	if err != nil {
		//credentials not found, using given username and password
		credentials = []byte(basicAuth(ctx.String("username"), ctx.String("password")))
	}
	return string(credentials)
}

func cacheCredentials(username, password string) {
	credentials := basicAuth(username, password)
	err := ioutil.WriteFile(CredentialsFile, []byte(credentials), 400)
	if err != nil {
		log.Printf("Caching credentials failed: %v", err)
	}
}

func checkDataDir() {
	if err := os.Mkdir(dataDir, 600); err == nil {
		log.Printf("Created cache directory at %s", dataDir)
	} else if !os.IsExist(err) {
		log.Printf("Error creating cache directory (%s): %v", dataDir, err)
	}
}

func cacheNotExist() bool {
	f, err := os.Open(MalCacheFile)
	defer f.Close()
	return os.IsNotExist(err)
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

func basicAuth(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

func reverseAnimeSlice(s []*mal.Anime) {
	last := len(s) - 1
	for i := 0; i < len(s)/2; i++ {
		s[i], s[last-i] = s[last-i], s[i]
	}
}

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Println("Error getting current user: %v", err)
		return ""
	}

	return usr.HomeDir
}
