package main

import (
	"encoding/json"
	"github.com/aqatl/mal/anilist"
	"github.com/aqatl/mal/oauth2"
	"os"
	"time"
	"fmt"
	"github.com/urfave/cli"
)

type AniList struct {
	Token oauth2.OAuthToken
	User  anilist.User
	List
}

type List []anilist.MediaListEntry

func (l List) GetMediaListById(listId int) *anilist.MediaListEntry {
	for i := 0; i < len(l); i++ {
		if l[i].ListId == listId {
			return &l[i]
		}
	}
	return nil
}

func (l List) GetMediaListByMalId(malId int) *anilist.MediaListEntry {
	for i := 0; i < len(l); i++ {
		if l[i].IdMal == malId {
			return &l[i]
		}
	}
	return nil
}

func sliceOfEntryTitles(entry *anilist.MediaListEntry) []string {
	alts := make([]string, 0, 3+len(entry.Synonyms))
	if t := entry.Title.English; t != "" {
		alts = append(alts, t)
	}
	if t := entry.Title.Native; t != "" {
		alts = append(alts, t)
	}
	if t := entry.Title.Romaji; t != "" {
		alts = append(alts, t)
	}
	alts = append(alts, entry.Synonyms...)
	return alts
}

func alGetList(al *AniList, status anilist.MediaListStatus) List {
	if status == anilist.All {
		return al.List
	} else {
		list := make(List, 0)
		for i := range al.List {
			if al.List[i].Status == status {
				list = append(list, al.List[i])
			}
		}
		return list
	}
}

func loadAniListFull(ctx *cli.Context) (al *AniList, entry *anilist.MediaListEntry, cfg *Config, err error) {
	al, err = loadAniList(ctx)
	if err != nil {
		return
	}
	cfg = LoadConfig()
	if cfg.ALSelectedID == 0 {
		fmt.Println("No entry selected")
	}
	entry = al.GetMediaListById(cfg.ALSelectedID)
	if entry == nil {
		err = fmt.Errorf("no entry found")
	}
	return
}

func loadAniList(ctx *cli.Context) (*AniList, error) {
	token, err := loadOAuthToken()
	if err != nil {
		return nil, err
	}
	al := &AniList{Token: token}

	if err := loadAniListUser(al); err != nil {
		return nil, err
	}
	if ctx.Bool("refresh") {
		err = fetchAniListAnimeLists(al)
	} else {
		err = loadAniListAnimeLists(al)
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

func loadAniListUser(al *AniList) error {
	if LoadJsonFile(AniListUserFile, &al.User) {
		return nil
	}
	err := anilist.QueryAuthenticatedUser(&al.User, al.Token)
	if err == anilist.InvalidToken {
		if al.Token, err = requestAniListToken(); err != nil {
			return err
		}
		err = anilist.QueryAuthenticatedUser(&al.User, al.Token)
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

func loadAniListAnimeLists(al *AniList) error {
	f, err := os.Open(AniListCacheFile)
	defer f.Close()
	if err == nil {
		err = json.NewDecoder(f).Decode(&al.List)
		return err
	}
	if !os.IsNotExist(err) {
		return err
	}
	return fetchAniListAnimeLists(al)
}

func fetchAniListAnimeLists(al *AniList) error {
	lists, err := anilist.QueryUserListsWaitAnimation(al.User.Id, al.Token)
	for i := range lists {
		al.List = append(al.List, lists[i].Entries...)
	}
	if err == anilist.InvalidToken {
		if al.Token, err = requestAniListToken(); err != nil {
			return err
		}
		lists, err = anilist.QueryUserListsWaitAnimation(al.User.Id, al.Token)
		for i := range lists {
			al.List = append(al.List, lists[i].Entries...)
		}
	}
	if err == nil {
		err = saveAniListAnimeLists(al)
	}
	return err
}

func saveAniListAnimeLists(al *AniList) error {
	f, err := os.Create(AniListCacheFile)
	defer f.Close()
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(&al.List)
	return err
}
