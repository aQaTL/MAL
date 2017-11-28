package mal

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
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

	AnimePage = BaseMALAddress + "/anime/%d" //%d - anime database ID
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

//credentials should be username + password encoded in the basic auth standard
func NewClient(credentials string) *Client {
	c := &Client{}

	credentialsBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(credentials, "Basic "))
	if err != nil {
		log.Printf("Decoding credentials failed: %v", err)
	}
	decodedCredentials := strings.Split(string(credentialsBytes), ":")
	c.Username = decodedCredentials[0]
	c.credentials = credentials

	return c
}

func VerifyCredentials(credentials string) bool {
	req := newRequest(VerifyCredentialsEndpoint, credentials, http.MethodGet)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Response error: %v", err)
		return false
	}

	if resp.StatusCode != 200 {
		log.Printf("Credentials verification status: %v", resp.Status)
	}

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

func (c *Client) AnimeList(status MyStatus) []*Anime {
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

//This works by scrapping the normal MAL website for given entry. It means that this method
//will stop working when they change something
func (c *Client) FetchDetails(entry *Anime) (*AnimeDetails, error) {
	url := fmt.Sprintf(AnimePage, entry.ID)

	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting response: %v", err)
	}

	reader, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	details := AnimeDetails{}

	details.JapaneseTitle = parseJapaneseTitle(reader)
	details.Related = parseRelated(reader)
	details.Characters = *parseCharacters(reader)
	details.ScoreVoters = parseScoreVoters(reader)

	synopsisNode := reader.Find("span[itemprop=description]")

	details.Synopsis = parseSynopsis(synopsisNode)
	details.Background = parseBackground(synopsisNode) //not working correctly

	spanDarkText := reader.Selection.Find("span[class=dark_text]")

	details.Premiered = parsePremiered(spanDarkText)
	details.Broadcast = parseBroadcast(spanDarkText)
	details.Producers = *parseProducers(spanDarkText)
	details.Licensors = *parseLicensors(spanDarkText)
	details.Studios = *parseStudios(spanDarkText)
	details.Source = parseSource(spanDarkText)
	details.Genres = *parseGenres(spanDarkText)
	details.Duration = parseDuration(spanDarkText)
	details.Rating = parseRating(spanDarkText)
	details.Score = parseScore(spanDarkText)
	details.Ranked = parseRanked(spanDarkText)
	details.Popularity = parsePopularity(spanDarkText)
	details.Members = parseMembers(spanDarkText)
	details.Favorites = parseFavorites(spanDarkText)

	//TODO parse Staff, OpeningThemes, EndingThemes

	return &details, nil
}
