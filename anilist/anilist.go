package anilist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aqatl/mal/oauth2"
	"github.com/pkg/errors"
)

const ApiEndpoint = "https://graphql.anilist.co"
const ALDomain = "https://anilist.co"

var InvalidToken = errors.New("Invalid token")

// TODO downloading only given list like watching/completed
func QueryUserLists(userId int, scoreFormat ScoreFormat, token oauth2.OAuthToken) ([]MediaListGroup, error) {
	vars := make(map[string]interface{})
	vars["userID"] = userId
	vars["scoreFormat"] = scoreFormat

	resp := struct {
		MediaListCollection `json:"MediaListCollection"`
	}{MediaListCollection{}}
	if err := gqlErrorsHandler(graphQLRequestParsed(queryUserAnimeList, vars, token, &resp)); err != nil {
		return nil, err
	}
	return resp.Lists, nil
}

func QueryAuthenticatedUser(user *User, token oauth2.OAuthToken) error {
	viewer := &struct {
		*User `json:"Viewer"`
	}{user}
	return gqlErrorsHandler(graphQLRequestParsed(queryAuthenticatedUser, nil, token, viewer))
}

func SaveMediaListEntry(entry *MediaListEntry, token oauth2.OAuthToken) error {
	vars := make(map[string]interface{})
	vars["listId"] = entry.ListId
	vars["mediaId"] = entry.Id
	vars["status"] = string(entry.Status)
	vars["progress"] = entry.Progress
	vars["score"] = entry.Score
	entryData := &struct {
		*MediaListEntry `json:"SaveMediaListEntry"`
	}{entry}
	return gqlErrorsHandler(graphQLRequestParsed(saveMediaListEntry, vars, token, entryData))
}

func AddMediaListEntry(id int, status MediaListStatus, token oauth2.OAuthToken) (
	MediaListEntry, error,
) {
	vars := make(map[string]interface{})
	vars["mediaId"] = id
	vars["status"] = string(status)
	entryData := &struct {
		MediaListEntry `json:"SaveMediaListEntry"`
	}{}
	err := gqlErrorsHandler(graphQLRequestParsed(addMediaListEntry, vars, token, entryData))
	return entryData.MediaListEntry, err
}

func QueryAiringSchedule(mediaId, episode int, token oauth2.OAuthToken) (AiringSchedule, error) {
	vars := make(map[string]interface{})
	vars["mediaId"] = mediaId
	vars["episode"] = episode
	data := &struct {
		AiringSchedule `json:"AiringSchedule"`
	}{AiringSchedule{}}
	err := gqlErrorsHandler(graphQLRequestParsed(queryAiringSchedule, vars, token, data))
	return data.AiringSchedule, err
}

func QueryAiringNotification(markRead bool, token oauth2.OAuthToken) (AiringNotification, error) {
	vars := make(map[string]interface{})
	vars["resetNotificationCount"] = markRead
	data := new(struct {
		AiringNotification `json:"Notification"`
	})
	err := gqlErrorsHandler(graphQLRequestParsed(queryAiringNotification, vars, token, data))
	return data.AiringNotification, err
}

func QueryAiringNotifications(page, perPage int, markRead bool, token oauth2.OAuthToken) (
	[]AiringNotification, error,
) {
	vars := make(map[string]interface{})
	vars["page"] = page
	vars["perPage"] = perPage
	vars["resetNotificationCount"] = markRead
	data := new(struct {
		Page struct {
			Notifications []AiringNotification `json:"notifications"`
		} `json:"Page"`
	})
	err := gqlErrorsHandler(graphQLRequestParsed(queryAiringNotifications, vars, token, data))
	return data.Page.Notifications, err
}

func DeleteMediaListEntry(entry *MediaListEntry, token oauth2.OAuthToken) error {
	vars := make(map[string]interface{})
	vars["id"] = entry.ListId
	data := new(struct {
		D struct {
			Deleted bool `json:"deleted"`
		} `json:"DeleteMediaListEntry"`
	})
	err := gqlErrorsHandler(graphQLRequestParsed(deleteMediaListEntry, vars, token, data))
	if err != nil {
		return err
	}
	if !data.D.Deleted {
		return fmt.Errorf("deletion unsuccessfull")
	}
	return nil
}

func Search(query string, page, perPage int, mtype MediaType, token oauth2.OAuthToken) ([]MediaFull, error) {
	vars := make(map[string]interface{})
	vars["page"] = page
	vars["perPage"] = perPage
	vars["search"] = query
	vars["type"] = mtype

	data := new(struct {
		Page struct {
			Media []MediaFull `json:"media"`
		} `json:"Page"`
	})
	err := gqlErrorsHandler(graphQLRequestParsed(queryMedia, vars, token, data))
	return data.Page.Media, err
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
		return fmt.Errorf("GraphQl Error (%d): %s\n%s",
			gqlErrs[0].Status, gqlErrs[0].Message, locations.String())
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

func graphQLRequestParsed(query string, vars map[string]interface{}, t oauth2.OAuthToken,
	x interface{}) ([]GqlError, error) {
	resp, err := graphQLRequest(query, vars, t)
	if resp != nil {
		defer resp.Body.Close()
	}
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
		// TODO include all error fields
		// TODO better error handling -> maybe typedef error array as QueryErrors?
		return respData.Errors, nil
	}
	return nil, nil
}

func graphQLRequestString(query string, vars map[string]interface{}, t oauth2.OAuthToken) (
	string, error,
) {
	resp, err := graphQLRequest(query, vars, t)
	if resp != nil {
		defer resp.Body.Close()
	}
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

	req, err := http.NewRequest(http.MethodPost, ApiEndpoint, &reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.Token)

	return http.DefaultClient.Do(req)
}

func ParseStatus(status string) MediaListStatus {
	switch strings.ToLower(status) {
	case "watching", "current":
		return Current
	case "planning", "plantowatch":
		return Planning
	case "completed":
		return Completed
	case "dropped":
		return Dropped
	case "paused", "onhold":
		return Paused
	case "repeating":
		return Repeating
	default:
		return All
	}
}
