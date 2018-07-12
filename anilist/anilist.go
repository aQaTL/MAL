package anilist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aqatl/mal/oauth2"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

//TODO Anilist support

const apiEndpoint = "https://graphql.anilist.co"

type AniList struct {
	Token oauth2.OAuthToken
	User  User
	MediaListCollection
}

var InvalidToken = errors.New("Invalid token")

//TODO downloading only given list like watching/completed
func (al *AniList) QueryUserLists() error {
	vars := make(map[string]interface{})
	vars["userID"] = al.User.Id

	resp := struct {
		MediaListCollection `json:"MediaListCollection"`
	}{MediaListCollection{}}
	if err := gqlErrorsHandler(graphQLRequestParsed(queryUserAnimeList, vars, &al.Token, &resp)); err != nil {
		return err
	}
	al.Lists = resp.Lists

	return nil
}

func (al *AniList) QueryAuthenticatedUser() error {
	viewer := &struct {
		*User `json:"Viewer"`
	}{&al.User}
	return gqlErrorsHandler(graphQLRequestParsed(queryAuthenticatedUser, nil, &al.Token, viewer))
}

func (al *AniList) SaveMediaListEntry(entry *MediaListEntry) error {
	vars := make(map[string]interface{})
	vars["listId"] = entry.ListId
	vars["mediaId"] = entry.Id
	vars["status"] = entry.Status
	vars["progress"] = entry.Progress
	vars["score"] = entry.Score
	entryData := &struct {
		*MediaListEntry `json:"SaveMediaListEntry"`
	}{entry}
	return gqlErrorsHandler(graphQLRequestParsed(saveMediaListEntry, vars, &al.Token, entryData))
}

func gqlErrorsHandler(gqlErrs []GqlError, err error) error {
	if err != nil {
		return err
	}
	for _, gqlErr := range gqlErrs {
		if gqlErr.Message == "Invalid token" {
			return InvalidToken
		}
	}
	if len(gqlErrs) > 0 {
		locations := strings.Builder{}
		for _, loc := range gqlErrs[0].Locations {
			locations.WriteString(fmt.Sprintf("Line %d column %d\n", loc.Line, loc.Column))
		}
		return fmt.Errorf("GraphQl Error (%d): %s\n%v",
			gqlErrs[0].Status, gqlErrs[0].Message, locations)
	}
	return nil
}

func printGqlErrs(gqlErrs []GqlError) {
	for _, gqlErr := range gqlErrs {
		fmt.Printf("GraphQl Error (%d): %s\n", gqlErr.Status, gqlErr.Message)
		for _, loc := range gqlErr.Locations {
			fmt.Printf("Line %d column %d\n", loc.Line, loc.Column)
		}
	}
}

func graphQLRequestParsed(query string, vars map[string]interface{}, t *oauth2.OAuthToken,
	x interface{}) ([]GqlError, error) {
	resp, err := graphQLRequest(query, vars, t)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	type responseData struct {
		Data   interface{}
		Errors []GqlError
	}
	respData := &responseData{Data: x}

	if err := json.NewDecoder(resp.Body).Decode(respData); err != nil {
		return nil, err
	}
	if len(respData.Errors) > 0 {
		//TODO include all error fields
		//TODO better error handling -> maybe typedef error array as QueryErrors?
		return respData.Errors, nil
	}
	return nil, nil
}

func graphQLRequestString(query string, vars map[string]interface{}, t *oauth2.OAuthToken) (
	string, error,
) {
	resp, err := graphQLRequest(query, vars, t)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	return string(data), err
}

func graphQLRequest(query string, vars map[string]interface{}, t *oauth2.OAuthToken) (
	*http.Response, error,
) {
	reqBody := bytes.Buffer{}
	err := json.NewEncoder(&reqBody).Encode(struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{query, vars})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, apiEndpoint, &reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.Token)

	return http.DefaultClient.Do(req)
}


func (c *MediaListCollection) GetMediaListById(listId int) *MediaListEntry {
	for i := 0; i < len(c.Lists); i++ {
		for j := 0; j < len(c.Lists[i].Entries); j++ {
			if c.Lists[i].Entries[j].ListId == listId {
				return &c.Lists[i].Entries[j]
			}
		}
	}
	return nil
}

func (c *MediaListCollection) GetMediaListByMalId(malId int) *MediaListEntry {
	for i := 0; i < len(c.Lists); i++ {
		for j := 0; j < len(c.Lists[i].Entries); j++ {
			if c.Lists[i].Entries[j].IdMal == malId {
				return &c.Lists[i].Entries[j]
			}
		}
	}
	return nil
}