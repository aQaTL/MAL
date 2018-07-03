package nyaa_scraper

import (
	"compress/gzip"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"net/url"
	"time"
	"strconv"
	"strings"
	"regexp"
)

type NyaaCategory struct {
	Major int
	Minor int
}

func (category NyaaCategory) String() string {
	return fmt.Sprintf("c=%d_%d", category.Major, category.Minor)
}

var (
	AllCategories = NyaaCategory{0, 0}

	Anime                     = NyaaCategory{1, 0}
	AnimeMusicVideo           = NyaaCategory{1, 1}
	AnimeEnglishTranslated    = NyaaCategory{1, 2}
	AnimeNonEnglishTranslated = NyaaCategory{1, 3}
	AnimeRaw                  = NyaaCategory{1, 4}

	Audio         = NyaaCategory{2, 0}
	AudioLossless = NyaaCategory{2, 1}
	AudioLossy    = NyaaCategory{2, 2}

	Literature                     = NyaaCategory{3, 0}
	LiteratureEnglishTranslated    = NyaaCategory{3, 1}
	LiteratureNonEnglishTranslated = NyaaCategory{3, 2}
	LiteratureRaw                  = NyaaCategory{3, 3}

	LiveAction                       = NyaaCategory{4, 0}
	LiveActionEnglishTranslated      = NyaaCategory{4, 1}
	LiveActionIdolOrPromotionalVideo = NyaaCategory{4, 2}
	LiveActionNonEnglishTranslated   = NyaaCategory{4, 3}
	LiveActionRaw                    = NyaaCategory{4, 4}

	Pictures         = NyaaCategory{5, 0}
	PicturesGraphics = NyaaCategory{5, 1}
	PicturesPhotos   = NyaaCategory{5, 2}

	Software             = NyaaCategory{6, 0}
	SoftwareApplications = NyaaCategory{6, 1}
	SoftwareGames        = NyaaCategory{6, 2}
)

type NyaaFilter uint8

const (
	NoFilter    NyaaFilter = iota
	NoRemakes
	TrustedOnly
)

func (filter NyaaFilter) String() string {
	return fmt.Sprintf("f=%d", filter)
}

type NyaaEntry struct {
	Category           NyaaCategory
	Title              string
	TorrentLink        string
	MagnetLink         string
	Size               string
	DateAdded          time.Time
	Seeders            int
	Leechers           int
	CompletedDownloads int
}

const nyaaQueryPattern = "https://nyaa.si/?%v&%v&p=%d&q=%s"

type NyaaResultPage struct {
	DisplayedFrom int
	DisplayedTo int
	DisplayedOutOf int

	Results []NyaaEntry
}

func Search(query string, category NyaaCategory, filter NyaaFilter) (NyaaResultPage, error) {
	return SearchSpecificPage(query, category, filter, 1)
}

func SearchSpecificPage(query string, category NyaaCategory, filter NyaaFilter, page int) (NyaaResultPage, error) {
	resultPage := NyaaResultPage{}

	//TODO check PathEscape vs QueryEscape
	address := fmt.Sprintf(nyaaQueryPattern, filter, category, page, url.PathEscape(query))
	respBody, err := doRequest(address)
	if err != nil {
		if respBody != nil {
			respBody.Close()
		}
		return resultPage, fmt.Errorf("request failed: %v", err)
	}
	defer respBody.Close()

	doc, err := goquery.NewDocumentFromReader(respBody)
	if err != nil {
		return resultPage, fmt.Errorf("error parsing response: %v", err)
	}

	rows := doc.Find(".torrent-list > tbody").Children()
	resultPage.Results = make([]NyaaEntry, rows.Size())

	rows.Each(func(i int, sel *goquery.Selection) {
		resultPage.Results[i] = *parseNyaaEntry(sel)
	})

	info := strings.TrimSpace(doc.Find("div .pagination-page-info").Text())

	re := regexp.MustCompile("[0-9]+")
	numbers := re.FindAllString(info, -1)

	if len(numbers) < 3 {
		return resultPage, fmt.Errorf("regexp failed (returned less than 3 resutls)")
	}
	resultPage.DisplayedFrom, _ = strconv.Atoi(numbers[0])
	resultPage.DisplayedTo, _ = strconv.Atoi(numbers[1])
	resultPage.DisplayedOutOf, _ = strconv.Atoi(numbers[2])

	return resultPage, nil
}

func parseNyaaEntry(sel *goquery.Selection) *NyaaEntry {
	entry := NyaaEntry{}

	currChild := sel.Children().First()
	category := currChild.Find("a").AttrOr("href", "")
	major, _ := strconv.Atoi(category[4:5])
	minor, _ := strconv.Atoi(category[6:7])
	entry.Category = NyaaCategory{major, minor}

	currChild = currChild.Next()
	entry.Title = currChild.Find("a").Last().AttrOr("title", "")

	currChild = currChild.Next()
	links := currChild.Find("a")
	entry.TorrentLink = "https://nyaa.si" + links.AttrOr("href", "")
	entry.MagnetLink = links.Next().AttrOr("href", "")

	currChild = currChild.Next()
	entry.Size = strings.TrimSpace(currChild.Text())

	currChild = currChild.Next()
	timestamp, _ := strconv.Atoi(currChild.AttrOr("data-timestamp", "0"))
	entry.DateAdded = time.Unix(int64(timestamp), 0)

	currChild = currChild.Next()
	seeders, _ := strconv.Atoi(currChild.Text())
	entry.Seeders = seeders

	currChild = currChild.Next()
	leechers, _ := strconv.Atoi(currChild.Text())
	entry.Leechers = leechers

	currChild = currChild.Next()
	completedDownloads, _ := strconv.Atoi(currChild.Text())
	entry.CompletedDownloads = completedDownloads

	return &entry
}

//Note: do not forget to close the returned ReadCloser
func doRequest(address string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("searching failed; server returned: %s", resp.Status)
	}

	return gzip.NewReader(resp.Body)
}
