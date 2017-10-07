package mal

import (
	"net/http"
	"log"
	"fmt"
	"encoding/base64"
	"strings"
	"encoding/xml"
	"io"
)

const (
	BaseMALAddress            = "https://myanimelist.net"
	ApiEndpoint               = BaseMALAddress + "/api"
	VerifyCredentialsEndpoint = ApiEndpoint + "/account/verify_credentials.xml"

	UserAnimeListEndpoint = BaseMALAddress + "/malappinfo.php?u=%s&status=%s&type=anime" //%s - username %s - status
)

type Client struct {
	Username    string
	credentials string

	ID          string `xml:"user_id"`
	Watching    int    `xml:"user_watching"`
	Completed   int    `xml:"user_completed"`
	OnHold      int    `xml:"user_onhold"`
	Dropped     int    `xml:"user_dropped"`
	PlanToWatch int    `xml:"user_plantowatch"`

	DaysSpentWatching float64 `xml:"user_days_spent_watching"`
}

func NewClient(credentials string) *Client {
	c := &Client{}

	if !verifyCredentials(credentials) {
		return nil
	}

	credentialsBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(credentials, "Basic "))
	if err != nil {
		log.Printf("Decoding credentials failed: %v", err)
	}
	decodedCredentials := strings.Split(string(credentialsBytes), ":")
	c.Username = decodedCredentials[0]
	c.credentials = credentials

	return c
}

func verifyCredentials(credentials string) bool {
	req := newRequest(VerifyCredentialsEndpoint, credentials, http.MethodGet)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Response error: %v", err)
		return false
	}
	log.Printf("Credentials verification status: %v", resp.Status)

	return resp.StatusCode == 200
}

func (c *Client) AnimeList(status myStatus) []*Anime {
	url := fmt.Sprintf(UserAnimeListEndpoint, c.Username, "all") //Anything other than `all` doesn't really work

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Printf("Request error: %v", err)
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("List list getting error: %v", err)
	}

	decoder := xml.NewDecoder(resp.Body)
	decoder.Strict = false

	list := make([]*Anime, 0)

	for t, err := decoder.Token(); err != io.EOF; t, err = decoder.Token() {
		if t, ok := t.(xml.StartElement); ok {
			switch t.Name.Local {
			case "myinfo":
				decoder.DecodeElement(&c, &t)
			case "anime":
				anime := new(Anime)
				decoder.DecodeElement(&anime, &t)
				if anime.MyStatus == status {
					list = append(list, anime)
				}
			}
		}
	}

	return list
}

func newRequest(url, credentials, method string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Printf("Request creation error: %v", err)
		return nil
	}
	req.Header.Add("Authorization", credentials)
	return req
}
