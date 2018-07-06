package nyaa_scraper

import (
	"compress/gzip"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type NyaaCategory struct {
	Name  string
	Major int
	Minor int
}

func (category NyaaCategory) QueryParam() string {
	return fmt.Sprintf("c=%d_%d", category.Major, category.Minor)
}

func (category NyaaCategory) String() string {
	return category.Name
}

var (
	AllCategories = NyaaCategory{"All categories", 0, 0}

	Anime                     = NyaaCategory{"All anime", 1, 0}
	AnimeMusicVideo           = NyaaCategory{"Anime music video", 1, 1}
	AnimeEnglishTranslated    = NyaaCategory{"Anime English translated", 1, 2}
	AnimeNonEnglishTranslated = NyaaCategory{"Anime non English translated", 1, 3}
	AnimeRaw                  = NyaaCategory{"Anime raw", 1, 4}

	Audio         = NyaaCategory{"All Audio", 2, 0}
	AudioLossless = NyaaCategory{"Audio lossless", 2, 1}
	AudioLossy    = NyaaCategory{"Audio lossy", 2, 2}

	Literature                     = NyaaCategory{"All literature", 3, 0}
	LiteratureEnglishTranslated    = NyaaCategory{"Literature English translated", 3, 1}
	LiteratureNonEnglishTranslated = NyaaCategory{"Literature non English translated", 3, 2}
	LiteratureRaw                  = NyaaCategory{"Literature raw", 3, 3}

	LiveAction                       = NyaaCategory{"Live action all", 4, 0}
	LiveActionEnglishTranslated      = NyaaCategory{"Live action English translated", 4, 1}
	LiveActionIdolOrPromotionalVideo = NyaaCategory{"Live action idol/promotional video", 4, 2}
	LiveActionNonEnglishTranslated   = NyaaCategory{"Live action non English translated", 4, 3}
	LiveActionRaw                    = NyaaCategory{"Live action raw", 4, 4}

	Pictures         = NyaaCategory{"Pictures all", 5, 0}
	PicturesGraphics = NyaaCategory{"Graphics", 5, 1}
	PicturesPhotos   = NyaaCategory{"Photos", 5, 2}

	Software             = NyaaCategory{"Software all", 6, 0}
	SoftwareApplications = NyaaCategory{"Applications", 6, 1}
	SoftwareGames        = NyaaCategory{"Games", 6, 2}
)

var Categories = []NyaaCategory{
	AllCategories,
	Anime,
	AnimeMusicVideo,
	AnimeEnglishTranslated,
	AnimeNonEnglishTranslated,
	AnimeRaw,
	Audio,
	AudioLossless,
	AudioLossy,
	Literature,
	LiteratureEnglishTranslated,
	LiteratureNonEnglishTranslated,
	LiteratureRaw,
	LiveAction,
	LiveActionEnglishTranslated,
	LiveActionIdolOrPromotionalVideo,
	LiveActionNonEnglishTranslated,
	LiveActionRaw,
	Pictures,
	PicturesGraphics,
	PicturesPhotos,
	Software,
	SoftwareApplications,
	SoftwareGames,
}

func GetNyaaCategory(major, minor int) NyaaCategory {
	for _, c := range Categories {
		if c.Major == major && c.Minor == minor {
			return c
		}
	}
	return NyaaCategory{"Unknown", major, minor}
}

type NyaaFilter struct {
	Name string
	Val uint8
}

var (
	NoFilter = NyaaFilter{"No filter", 0}
	NoRemakes = NyaaFilter{"No remakes", 1}
	TrustedOnly = NyaaFilter{"Trusted only", 2}
)

func (filter NyaaFilter) QueryParam() string {
	return fmt.Sprintf("f=%d", filter.Val)
}

func (filter NyaaFilter) String() string {
	return filter.Name
}

var Filters = []NyaaFilter{
	NoFilter,
	NoRemakes,
	TrustedOnly,
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

const nyaaQueryPattern = "https://nyaa.si/?%s&%s&p=%d&q=%s"

type NyaaResultPage struct {
	DisplayedFrom  int
	DisplayedTo    int
	DisplayedOutOf int

	Results []NyaaEntry
}

func Search(query string, category NyaaCategory, filter NyaaFilter) (NyaaResultPage, error) {
	return SearchSpecificPage(query, category, filter, 1)
}

func SearchSpecificPage(query string, category NyaaCategory, filter NyaaFilter, page int) (NyaaResultPage, error) {
	resultPage := NyaaResultPage{}

	address := fmt.Sprintf(nyaaQueryPattern, filter.QueryParam(), category.QueryParam(),
		page, url.QueryEscape(query))
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
	entry.Category = GetNyaaCategory(major, minor)

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
