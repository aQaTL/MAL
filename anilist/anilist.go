package anilist

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aqatl/mal/oauth2"
	"io/ioutil"
	"net/http"
)

//TODO Anilist support

const apiEndpoint = "https://graphql.anilist.co"

type AniList struct {
	Token oauth2.OAuthToken
	User User
}

type User struct {
	Id                      int       `json:"id"`
	Name                    string    `json:"name"`
	About                   string    `json:"about"`
	BannerImage             string    `json:"bannerImage"`
	Stats                   UserStats `json:"stats"`
	UnreadNotificationCount int       `json:"unreadNotificationCount"`
	SiteUrl                 string    `json:"siteUrl"`
	DonatorTier             int       `json:"donatorTier"`
	ModeratorStatus         string    `json:"moderatorStatus"`
	UpdatedAt               int       `json:"updatedAt"`
}

type UserStats struct {
	WatchedTime int `json:"watchedTime"`
}

func QueryAuthenticatedUser(user *User, token oauth2.OAuthToken) error {
	query := `
query {
	Viewer {
		id
		name
		about
		bannerImage
		stats {
			watchedTime
		}
		unreadNotificationCount
		siteUrl
		donatorTier
		moderatorStatus
		updatedAt
	}
}`
	viewer := &struct {
		*User `json:"Viewer"`
	}{user}
	return graphQLRequestParsed(query, nil, token, viewer)
	return nil
}

//TODO downloading only given list like watching/completed
func (al *AniList) QueryUserList() error {
	vars := make(map[string]interface{})
	vars["userID"] = al.User.Id

	response, err := graphQLRequestString(queryUserAnimeList, vars, al.Token)
	if err != nil {
		return err
	}

	//TODO parsing response
	fmt.Println(response)

	return nil
}

func graphQLRequestParsed(query string, vars map[string]interface{}, t oauth2.OAuthToken,
	x interface{}) error {
	resp, err := graphQLRequest(query, vars, t)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	type responseData struct {
		Data   interface{}
		Errors []struct {
			Message string
		}
	}
	respData := &responseData{Data: x}

	if err := json.NewDecoder(resp.Body).Decode(respData); err != nil {
		return err
	}
	if len(respData.Errors) > 0 {
		//TODO include all error fields
		//TODO better error handling -> maybe typedef error array as QueryErrors?
		errs := ""
		for _, err := range respData.Errors {
			errs += err.Message
		}
		return errors.New(errs)
	}
	return nil
}

func graphQLRequestString(query string, vars map[string]interface{}, t oauth2.OAuthToken) (
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

func graphQLRequest(query string, vars map[string]interface{}, t oauth2.OAuthToken) (
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
