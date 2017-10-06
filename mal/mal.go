package mal

import (
	"net/http"
	"log"
)

const (
	ApiEndpoint               = "https://myanimelist.net/api"
	VerifyCredentialsEndpoint = ApiEndpoint + "/account/verify_credentials.xml"
)

type status int

const (
	Watching    status = iota
	Completed
	OnHold
	Dropped
	PlanToWatch
)

type Client struct {
	Username string
	password string
}

func NewClient(credentials string) *Client {
	c := &Client{}

	if !verifyCredentials(credentials) {
		return nil
	}

	return c
}

func verifyCredentials(credentials string) bool {
	req, err := http.NewRequest("GET", VerifyCredentialsEndpoint, nil)
	if err != nil {
		log.Printf("Request creation error: %v", err)
		return false
	}
	req.Header.Add("Authorization", credentials)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Response error: %v", err)
		return false
	}
	log.Printf("Credentials verification status: %v", resp.Status)

	return resp.StatusCode == 200
}

func (c *Client) GetAnimeList(status status) []Anime {
	log.Printf("%v", status)

	return nil
}
