package main

import (
	"encoding/json"
	"github.com/aqatl/mal/anilist"
	"github.com/aqatl/mal/oauth2"
	"github.com/urfave/cli"
	"os"
	"time"
	"fmt"
)

func openAniList(ctx *cli.Context) error {
	al, err := loadAniList()
	if err != nil {
		return err
	}

	if err = al.QueryUserLists(); err != nil {
		return err
	}

	counter := 0
	for _, listGroup := range al.Lists {
		for _, entry := range listGroup.Entries {
			fmt.Println(entry.Title.Romaji)
			counter++
		}
	}

	fmt.Println(counter)

	return nil
}

func loadAniList() (*anilist.AniList, error) {
	token, err := loadOAuthToken()
	if err != nil {
		return nil, err
	}

	user, err := loadAniListUser(token)
	if err != nil {
		return nil, err
	}

	al := anilist.AniList{Token: token, User: *user}
	return &al, nil
}

func loadOAuthToken() (oauth2.OAuthToken, error) {
	token, err := loadCachedOAuthToken()
	_, isInvalidToken := err.(invalidToken)
	if err != nil {
		if !isInvalidToken && !os.IsNotExist(err) {
			return token, err
		} else {
			token, err = requestAniListToken()
			if err == nil {
				err = saveOAuthToken(token)
			}
			return token, err
		}
	}
	return token, nil
}

type invalidToken string

func (t invalidToken) Error() string {
	return string(t)
}

const (
	emptyToken   = invalidToken("empty token")
	tokenExpired = invalidToken("token expired")
)

func loadCachedOAuthToken() (oauth2.OAuthToken, error) {
	token := oauth2.OAuthToken{}

	f, err := os.Open(AniListCredsFile)
	defer f.Close()
	if err != nil {
		return token, err
	}

	if err = json.NewDecoder(f).Decode(&token); err != nil {
		return token, err
	}
	if token.Token == "" {
		return token, emptyToken
	}
	if token.ExpireDate.Before(time.Now()) {
		return token, tokenExpired
	}

	return token, err
}

func saveOAuthToken(token oauth2.OAuthToken) error {
	f, err := os.Create(AniListCredsFile)
	defer f.Close()
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(&token)
	return err
}

func requestAniListToken() (oauth2.OAuthToken, error) {
	return oauth2.OAuthImplicitGrantAuth(
		"https://anilist.co/api/v2/oauth/authorize",
		LoadConfig().BrowserPath,
		743,
		42505,
	)
}

func loadAniListUser(token oauth2.OAuthToken) (*anilist.User, error) {
	user := &anilist.User{}
	f, err := os.Open(AniListUserFile)
	if err == nil {
		err = json.NewDecoder(f).Decode(user)
		return user, err
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	err = anilist.QueryAuthenticatedUser(user, token)
	if err == nil {
		err = saveAniListUser(user)
	}
	return user, err
}

func saveAniListUser(user *anilist.User) error {
	f, err := os.Create(AniListUserFile)
	defer f.Close()
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(&user)
	return err
}
