package mal

import (
	"strings"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"time"
	"github.com/fatih/color"
	"bufio"
	"os"
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

func DoFuncWithWaitAnimation(text string, f func()) {
	done := make(chan struct{})
	go func() {
		f()
		done <- struct{}{}
	}()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	green := color.New(color.FgHiGreen).FprintFunc()
	clockStates := [...]string{"-", "\\", "|", "/"}
	currClockState := 0

	stdout := bufio.NewWriter(os.Stdout)
	for {
		select {
		case <-ticker.C:
			green(stdout, fmt.Sprintf("\r%s %s", text, clockStates[currClockState]))
			stdout.Flush()
			currClockState = (currClockState + 1) % len(clockStates)
		case <-done:
			fmt.Fprintf(stdout, "\r%s\r", strings.Repeat(" ", len(text)+4))
			stdout.Flush()
			return
		}
	}
}

func FetchDetailsWithAnimation(c *Client, entry *Anime) (*AnimeDetails, error) {
	var details *AnimeDetails
	var err error
	DoFuncWithWaitAnimation("Fetching details", func() {
		details, err = c.FetchDetails(entry)
	})
	return details, err

}