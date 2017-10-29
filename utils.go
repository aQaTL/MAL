package main

import (
	"bufio"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/aqatl/mal/mal"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"
	"syscall"
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
		credentials, err := ioutil.ReadFile(CredentialsFile)
		if err != nil {
			log.Printf("Failed to load credentials: %v", err)
			return ""
		}
		return string(credentials)
	}
}

func saveCredentials(credentials string) {
	err := ioutil.WriteFile(CredentialsFile, []byte(credentials), 400)
	if err != nil {
		log.Printf("Caching credentials failed: %v", err)
	}
}

func checkDataDir() {
	if err := os.Mkdir(dataDir, os.ModePerm); err == nil {
		log.Printf("Created cache directory at %s", dataDir)
	} else if !os.IsExist(err) {
		log.Printf("Error creating cache directory (%s): %v", dataDir, err)
	}
}

func loadList(c *mal.Client, ctx *cli.Context) []*mal.Anime {
	var list []*mal.Anime
	if ctx.Bool("refresh") || cacheNotExist() {
		list = c.AnimeList(mal.All)
	} else {
		list = loadCachedList()
	}
	return list
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
