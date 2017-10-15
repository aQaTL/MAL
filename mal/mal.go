package mal

import (
	"net/http"
	"log"
	"fmt"
	"encoding/base64"
	"strings"
	"encoding/xml"
	"io"
	"text/template"
	"bytes"
	"net/url"
	"io/ioutil"
)

const (
	BaseMALAddress            = "https://myanimelist.net"
	ApiEndpoint               = BaseMALAddress + "/api"
	VerifyCredentialsEndpoint = ApiEndpoint + "/account/verify_credentials.xml"
)

//For using as a printf format
const (
	UpdateEndpoint = ApiEndpoint + "/animelist/update/%d.xml" //%d - anime database ID

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

func newRequest(url, credentials, method string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Printf("Request creation error: %v", err)
		return nil
	}
	req.Header.Add("Authorization", credentials)
	return req
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
				if anime.MyStatus == status || status == All {
					list = append(list, anime)
				}
			}
		}
	}

	return list
}

func (c *Client) Update(entry *Anime) bool {
	buf := &bytes.Buffer{}

	template.Must(
		template.New("animeXML").
			Parse(AnimeXMLTemplate)).
		Execute(buf, entry)

	payload := url.Values{}
	payload.Set("data", buf.String())

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf(UpdateEndpoint, entry.ID),
		strings.NewReader(payload.Encode()),
	)

	if err != nil {
		log.Printf("Error creating http request: %v", err)
		return false
	}
	req.Header.Set("Authorization", c.credentials)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error getting response for %d - %s update: %v", entry.ID, entry.Title, err)
		return false
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}
	body := string(bodyBytes)

	if body != "Updated" || resp.StatusCode != 200 {
		log.Printf("Body: %v\nStatus: %s", body, resp.Status)
		return false
	}
	return true
}
