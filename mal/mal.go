package mal

import (
	"net/http"
	"log"
)

const (
	ApiEndpoint               = "https://myanimelist.net/api"
	VerifyCredentialsEndpoint = ApiEndpoint + "/account/verify_credentials.xml"
)

type Client struct {
	Username string
	password string
}

func NewClient(username, password string) *Client {
	c := &Client{}

	if !verifyCredentials(username, password) {
		return nil
	}

	return c
}

func verifyCredentials(username, password string) bool {
	req, err := http.NewRequest("GET", VerifyCredentialsEndpoint, nil)
	if err != nil {
		log.Printf("Request creation error: %v", err)
		return false
	}
	req.SetBasicAuth(username, password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Response error: %v", err)
		return false
	}
	log.Printf("Credentials verification status: %v", resp.Status)

	return resp.StatusCode == 200
}
