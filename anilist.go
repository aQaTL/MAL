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

	user, err := loadAniListUser(token)
	if err != nil {
		return nil, err
	}

	al := &anilist.AniList{Token: token, User: *user}

	err = loadAniListAnimeLists(al)
	if err != nil {
		return nil, err
	}

	return al, nil
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
	LoadJsonFile(AniListCredsFile, &token)

	if token.Token == "" {
		return token, emptyToken
	}
	if token.ExpireDate.Before(time.Now()) {
		return token, tokenExpired
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
	if LoadJsonFile(AniListUserFile, &user) {
		return user, nil
	}
	err := anilist.QueryAuthenticatedUser(user, token)
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
