package mal

import (
	"strings"
	"fmt"
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