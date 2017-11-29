package mal

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"log"
	"strconv"
	"strings"
)

func parseRating(spanDarkText *goquery.Selection) string {
	return strings.TrimSpace(spanDarkText.
		FilterFunction(isTextEqualFilterFunc("Rating:")).
		Nodes[0].
		NextSibling.
		Data)
}

func parseDuration(spanDarkText *goquery.Selection) string {
	return strings.TrimSpace(spanDarkText.
		FilterFunction(isTextEqualFilterFunc("Duration:")).
		Nodes[0].
		NextSibling.
		Data)
}

func parseGenres(spanDarkText *goquery.Selection) *[]string {
	genres := make([]string, 0)
	spanDarkText.FilterFunction(isTextEqualFilterFunc("Genres:")).
		Siblings().
		Each(func(i int, s *goquery.Selection) {
			genres = append(genres, s.Text())
		})
	return &genres
}

func parseSource(spanDarkText *goquery.Selection) string {
	return strings.TrimSpace(spanDarkText.
		FilterFunction(isTextEqualFilterFunc("Source:")).
		Nodes[0].
		NextSibling.
		Data)
}

func parseStudios(spanDarkText *goquery.Selection) *[]string {
	studios := make([]string, 0)

	spanDarkText.FilterFunction(isTextEqualFilterFunc("Studios:")).
		Siblings().
		Each(func(i int, s *goquery.Selection) {
			studios = append(studios, s.Text())
		})

	return &studios
}

func parseLicensors(spanDarkText *goquery.Selection) *[]string {
	licensors := make([]string, 0)

	spanDarkText.FilterFunction(isTextEqualFilterFunc("Licensors:")).
		Siblings().
		Each(func(i int, s *goquery.Selection) {
			licensors = append(licensors, s.Text())
		})

	return &licensors
}

func parseProducers(spanDarkText *goquery.Selection) *[]string {
	producers := make([]string, 0)

	spanDarkText.FilterFunction(isTextEqualFilterFunc("Producers:")).
		Siblings().
		Each(func(i int, s *goquery.Selection) {
			producers = append(producers, s.Text())
		})

	return &producers
}

func parseBroadcast(spanDarkText *goquery.Selection) string {
	return strings.TrimSpace(spanDarkText.
		FilterFunction(isTextEqualFilterFunc("Broadcast:")).
		Nodes[0].
		NextSibling.
		Data)
}

func parsePremiered(spanDarkText *goquery.Selection) string {
	return spanDarkText.
		FilterFunction(isTextEqualFilterFunc("Premiered:")).
		Next().
		Text()
}

func parseBackground(synopsisNode *goquery.Selection) string {
	buf := bytes.Buffer{}
	for _, a := range synopsisNode.NextAll().Nodes {
		for sib := a.NextSibling; sib != nil; sib = sib.NextSibling {
			if sib.Type == html.TextNode {
				buf.WriteString(sib.Data)
			}
		}
	}
	return buf.String()
}

func parseSynopsis(synopsisNode *goquery.Selection) string {
	return synopsisNode.Text()
}

func parseJapaneseTitle(reader *goquery.Document) string {
	return strings.TrimSpace(reader.Find("div .spaceit_pad span").
		FilterFunction(isTextEqualFilterFunc("Japanese:")).
		Nodes[0].
		NextSibling.
		Data)
}

func parseRelated(reader *goquery.Document) []Related {
	relateds := make([]Related, 0)

	reader.Selection.Find(".anime_detail_related_anime tr").Each(
		func(i int, s *goquery.Selection) {
			relation := s.Find("td").First().Text()
			relation = relation[:len(relation)-1]

			link := s.Find("a")
			title := link.Text()
			url, _ := link.Attr("href")
			url = BaseMALAddress + url

			related := Related{relation, title, url}
			relateds = append(relateds, related)
		})

	return relateds
}

func parseCharacters(reader *goquery.Document) *[]Character {
	characters := make([]Character, 0)

	reader.Selection.
		Find("div .detail-characters-list").
		First().
		Find("table").
		FilterFunction(func(i int, s *goquery.Selection) bool {
			return i%2 == 0
		}).
		Each(func(i int, s *goquery.Selection) {
			c := Character{}
			tdNodes := s.Find("td").Next()

			names := [2]string{}
			tdNodes.Find("a").Each(func(i int, s *goquery.Selection) {
				if i < len(names) {
					names[i] = strings.TrimSpace(s.Text())
				}
			})
			c.Name = names[0]
			c.VoiceActor = names[1]

			roleAndActorOrigin := [2]string{}
			tdNodes.Find("small").Each(func(i int, s *goquery.Selection) {
				if i < len(roleAndActorOrigin) {
					roleAndActorOrigin[i] = strings.TrimSpace(s.Text())
				}
			})
			c.Role = roleAndActorOrigin[0]
			c.VoiceActorOrigin = roleAndActorOrigin[1]

			characters = append(characters, c)
		})

	return &characters
}

func parseScore(spanDarkText *goquery.Selection) float64 {
	score, err := strconv.ParseFloat(
		spanDarkText.FilterFunction(isTextEqualFilterFunc("Score:")).
			Next().
			Text(),
		64)
	if err != nil {
		log.Printf("error parsing score: %v", err)
		return 0
	}

	return score
}

func parseScoreVoters(reader *goquery.Document) int {
	voters, err := strconv.Atoi(strings.Replace(
		reader.Selection.
			Find("span[itemprop=ratingCount]").Nodes[0].FirstChild.Data,
		",",
		"",
		-1))
	if err != nil {
		log.Printf("error parsing ScoreVoters: %v", err)
		return 0
	}

	return voters
}

func parseRanked(spanDarkText *goquery.Selection) int {
	ranked, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(
		spanDarkText.
			FilterFunction(isTextEqualFilterFunc("Ranked:")).
			Nodes[0].
			NextSibling.
			Data),
		"#"))
	if err != nil {
		log.Printf("error parsing Ranked: %v", err)
		return 0
	}

	return ranked
}

func parsePopularity(spanDarkText *goquery.Selection) int {
	popularity, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(
		spanDarkText.
			FilterFunction(isTextEqualFilterFunc("Popularity:")).
			Nodes[0].
			NextSibling.
			Data),
		"#"))
	if err != nil {
		log.Printf("error parsing Popularity: %v", err)
		return 0
	}

	return popularity
}

func parseMembers(spanDarkText *goquery.Selection) int {
	members, err := strconv.Atoi(strings.Replace(strings.TrimSpace(
		spanDarkText.FilterFunction(isTextEqualFilterFunc("Members:")).
			Nodes[0].
			NextSibling.
			Data),
		",",
		"",
		-1))
	if err != nil {
		log.Printf("error parsing Members: %v", err)
		return 0
	}

	return members
}

func parseFavorites(spanDarkText *goquery.Selection) int {
	favorites, err := strconv.Atoi(strings.Replace(strings.TrimSpace(
		spanDarkText.FilterFunction(isTextEqualFilterFunc("Favorites:")).
			Nodes[0].
			NextSibling.
			Data),
		",",
		"",
		-1))
	if err != nil {
		log.Printf("error parsing Favorites: %v", err)
		return 0
	}

	return favorites
}

func parseStaff(reader *goquery.Document) *[]Staff {
	staff := make([]Staff, 0)

	reader.Selection.
		Find("div .detail-characters-list").
		Eq(1).
		Find("table").
		Each(func(i int, s *goquery.Selection) {
			name := strings.TrimSpace(s.Find("a").Last().Text())
			position := strings.TrimSpace(s.Find("small").Text())
			staff = append(staff, Staff{name, position})
		})

	return &staff
}

func parseOpeningThemes(reader *goquery.Document) *[]string {
	openingThemes := make([]string, 0)

	reader.Selection.
		Find(".opnening span").
		Each(func(i int, s *goquery.Selection) {
			song := strings.TrimSpace(s.Text())
			openingThemes = append(openingThemes, song)
		})

	return &openingThemes
}

func parseEndingThemes(reader *goquery.Document) *[]string {
	endingThemes := make([]string, 0)

	reader.Selection.
		Find(".ending span").
		Each(func(i int, s *goquery.Selection) {
			song := strings.TrimSpace(s.Text())
			endingThemes = append(endingThemes, song)
		})

	return &endingThemes
}
