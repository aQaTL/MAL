package anilist

import "strings"

type User struct {
	Id                      int       `json:"id"`
	Name                    string    `json:"name"`
	About                   string    `json:"about"`
	BannerImage             string    `json:"bannerImage"`
	Stats                   UserStats `json:"stats"`
	UnreadNotificationCount int       `json:"unreadNotificationCount"`
	SiteUrl                 string    `json:"siteUrl"`
	DonatorTier             int       `json:"donatorTier"`
	ModeratorStatus         string    `json:"moderatorStatus"`
	UpdatedAt               int       `json:"updatedAt"`
}

type UserStats struct {
	WatchedTime int `json:"watchedTime"`
}

type MediaListCollection struct {
	Lists []MediaListGroup `json:"lists"`
}

type MediaListGroup struct {
	Entries              []MediaListEntry `json:"entries"`
	Name                 string           `json:"name"`
	IsCustomList         bool             `json:"isCustomList"`
	IsSplitCompletedList bool             `json:"isSplitCompletedList"`
	Status               MediaListStatus  `json:"status"`
}

type MediaListEntry struct {
	ListId    int             `json:"id"`
	Status    MediaListStatus `json:"status"`
	Score     int             `json:"score"`
	Progress  int             `json:"progress"`
	UpdatedAt int             `json:"updatedAt"`

	Media `json:"media"`
}

type Media struct {
	Id                int            `json:"id"`
	IdMal             int            `json:"idMal"`
	Title             MediaTitle     `json:"title"`
	Type              string         `json:"type"`
	Format            string         `json:"format"`
	Status            string         `json:"status"`
	Description       string         `json:"description"`
	StartDate         FuzzyDate      `json:"startDate"`
	EndDate           FuzzyDate      `json:"endDate"`
	Season            string         `json:"season"`
	Episodes          int            `json:"episodes"`
	Duration          int            `json:"duration"`
	Source            string         `json:"source"`
	Genres            []string       `json:"genres"`
	Synonyms          []string       `json:"synonyms"`
	AverageScore      int            `json:"averageScore"`
	Popularity        int            `json:"popularity"`
	NextAiringEpisode AiringSchedule `json:"nextAiringEpisode"`
}

type MediaTitle struct {
	Romaji        string `json:"romaji"`
	English       string `json:"english"`
	Native        string `json:"native"`
	UserPreferred string `json:"userPreferred"`
}

type FuzzyDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

type AiringSchedule struct {
	Id              int `json:"id"`
	AiringAt        int `json:"airingAt"`
	TimeUntilAiring int `json:"timeUntilAiring"`
	Episode         int `json:"episode"`
}

type GqlError struct {
	Message   string     `json:"message"`
	Status    int        `json:"status"`
	Locations []Location `json:"locations"`
}

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type MediaListStatus string

const (
	All       MediaListStatus = ""
	Current   MediaListStatus = "CURRENT"
	Planning  MediaListStatus = "PLANNING"
	Completed MediaListStatus = "COMPLETED"
	Dropped   MediaListStatus = "DROPPED"
	Paused    MediaListStatus = "PAUSED"
	Repeating MediaListStatus = "REPEATING"
)

func (status MediaListStatus) String() string {
	if status == All {
		return ""
	} else if status == Current {
		return "Watching"
	} else {
		return string(status[0]) + strings.ToLower(string(status[1:]))
	}
}
