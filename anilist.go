package main

import (
	"encoding/json"
	"github.com/aqatl/mal/anilist"
	"github.com/aqatl/mal/oauth2"
	"os"
	"time"
)

func loadAniList() (*anilist.AniList, error) {
	token, err := loadOAuthToken()
	if err != nil {
		return nil, err
	}
	al := &anilist.AniList{Token: token}

	if err := loadAniListUser(al); err != nil {
		return nil, err
	}
	if err = loadAniListAnimeLists(al); err != nil {
		return nil, err
	}

	return al, nil
}

func loadOAuthToken() (oauth2.OAuthToken, error) {
	token, err := loadCachedOAuthToken()
	if err != nil {
		if err == anilist.InvalidToken || os.IsNotExist(err) {
			token, err = requestAniListToken()
		}
	}
	return token, err
}

func loadCachedOAuthToken() (oauth2.OAuthToken, error) {
	token := oauth2.OAuthToken{}
	LoadJsonFile(AniListCredsFile, &token)
	if token.Token == "" || token.ExpireDate.Before(time.Now()) {
		return token, anilist.InvalidToken
	}
	return token, nil
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

func requestAniListToken() (token oauth2.OAuthToken, err error) {
	token, err = oauth2.OAuthImplicitGrantAuth(
		"https://anilist.co/api/v2/oauth/authorize",
		LoadConfig().BrowserPath,
		743,
		42505,
	)
	if err != nil {
		return
	}
	err = saveOAuthToken(token)
	return
}

func loadAniListUser(al *anilist.AniList) error {
	if LoadJsonFile(AniListUserFile, &al.User) {
		return nil
	}
	err := al.QueryAuthenticatedUser()
	if err == anilist.InvalidToken {
		if al.Token, err = requestAniListToken(); err != nil {
			return err
		}
		err = al.QueryAuthenticatedUser()
	}
	if err == nil {
		err = saveAniListUser(&al.User)
	}
	return err
}

func saveAniListUser(user *anilist.User) error {
	f, err := os.Create(AniListUserFile)
	defer f.Close()
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(user)
	return err
}

func loadAniListAnimeLists(al *anilist.AniList) error {
	f, err := os.Open(AniListCacheFile)
	defer f.Close()
	if err == nil {
		err = json.NewDecoder(f).Decode(&al.Lists)
		return err
	}
	if !os.IsNotExist(err) {
		return err
	}
	err = al.QueryUserLists()
	if err == anilist.InvalidToken {
		if al.Token, err = requestAniListToken(); err != nil {
			return err
		}
		err = al.QueryUserLists()
	}
	if err == nil {
		err = saveAniListAnimeLists(al)
	}
	return err
}

func saveAniListAnimeLists(al *anilist.AniList) error {
	f, err := os.Create(AniListCacheFile)
	defer f.Close()
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(&al.Lists)
	return err
}
