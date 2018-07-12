package anilist

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
	Entries              []MediaList `json:"entries"`
	Name                 string      `json:"name"`
	IsCustomList         bool        `json:"isCustomList"`
	IsSplitCompletedList bool        `json:"isSplitCompletedList"`
	Status               string      `json:"status"`
}

type MediaList struct {
	Status   string `json:"status"`
	Score    int    `json:"score"`
	Progress int    `json:"progress"`
	Media    `json:"media"`
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
	Duration          int            `json:"duration"`
	Source            string         `json:"source"`
	UpdatedAt         int            `json:"updatedAt"`
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
