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
	Lists []MediaListGroup
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

type MediaListCollection struct {
	Lists []MediaListGroup `json:"lists"`
}

type MediaListGroup struct {
	Entries              []MediaList `json:"entries"`
	Name                 string      `json:"name"`
	IsCustomList         bool        `json:"isCustomList"`
	IsSplitCompletedList bool        `json:"isSplitCompletedList"`
	Status               string      `json:"status"`
}

type MediaList struct {
	Status   string `json:"status"`
	Score    int    `json:"score"`
	Progress int    `json:"progress"`
	Media    `json:"media"`
}

type Media struct {
	Id                int            `json:"id"`
	IdMal             int            `json:"idMal"`
	Title             MediaTitle     `json:"title"`
	Type              string         `json:"type"`
	Format            string         `json:"format"`
	Status            string         `json:"status"`
	Description       string         `json:"description"`
	StartDate         FuzzyDate      `json:"startDate"`
	EndDate           FuzzyDate      `json:"endDate"`
	Season            string         `json:"season"`
	Duration          int            `json:"duration"`
	Source            string         `json:"source"`
	UpdatedAt         int            `json:"updatedAt"`
	Genres            []string       `json:"genres"`
	Synonyms          []string       `json:"synonyms"`
	AverageScore      int            `json:"averageScore"`
	Popularity        int            `json:"popularity"`
	NextAiringEpisode AiringSchedule `json:"nextAiringEpisode"`
}

type MediaTitle struct {
	Romaji        string `json:"romaji"`
	English       string `json:"english"`
	Native        string `json:"native"`
	UserPreferred string `json:"userPreferred"`
}

type FuzzyDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

type AiringSchedule struct {
	Id              int `json:"id"`
	AiringAt        int `json:"airingAt"`
	TimeUntilAiring int `json:"timeUntilAiring"`
	Episode         int `json:"episode"`
}

type GqlError struct {
	Message   string     `json:"message"`
	Status    int        `json:"status"`
	Locations []Location `json:"locations"`
}

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
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

func QueryAuthenticatedUser(user *User, token *oauth2.OAuthToken) error {
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
	return gqlErrorsHandler(graphQLRequestParsed(query, nil, token, viewer))
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
