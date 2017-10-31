package mal

import "strings"

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
