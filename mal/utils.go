package mal

import (
	"strings"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/aqatl/cliwait"
)

func ParseStatus(status string) MyStatus {
	switch strings.ToLower(status) {
	case "watching":
		return Watching
	case "completed":
		return Completed
	case "onhold":
		return OnHold
	case "dropped":
		return Dropped
	case "plantowatch":
		return PlanToWatch
	default:
		 return All
	}
}

func ParseScore(score int) (AnimeScore, error) {
	if score < 0 || score > 10 {
		return 0, fmt.Errorf("score can not be outside of the 0-10 rage")
	}

	return AnimeScore(score), nil
}

func isTextEqualFilterFunc(text string) func(i int, s *goquery.Selection) bool {
	return func(i int, s *goquery.Selection) bool {
		return s.Text() == text
	}
}

func FetchDetailsWithAnimation(c *Client, entry *Anime) (*AnimeDetails, error) {
	var details *AnimeDetails
	var err error
	cliwait.DoFuncWithWaitAnimation("Fetching details", func() {
		details, err = c.FetchDetails(entry)
	})
	return details, err

}

func UpdateEntryWithAnimation(c *Client, entry *Anime) (error) {
	var err error
	cliwait.DoFuncWithWaitAnimation("Updating entry", func() {
		err = c.Update(entry)
	})
	return err
}