package mal

import "sort"

type animeType int

const (
	Tv      animeType = iota + 1
	Ova
	Movie
	Special
	Ona
	Music
)

type animeStatus int

const (
	CurrentlyAiring animeStatus = iota + 1
	FinishedAiring
	NotYetAired
)

type AnimeScore int

const (
	NotRatedYet AnimeScore = iota
	Appalling
	Horrible
	VeryBad
	Bad
	Average
	Fine
	Good
	VeryGood
	Great
	Masterpiece
)

type MyStatus int

const (
	All         MyStatus = iota
	Watching
	Completed
	OnHold
	Dropped
	PlanToWatch MyStatus = 6 //Apparently MAL stores this as 6
)

type Anime struct {
	ID          int         `xml:"series_animedb_id"`
	Title       string      `xml:"series_title"`
	Synonyms    string      `xml:"series_synonyms"`
	Type        animeType   `xml:"series_type"`
	Episodes    int         `xml:"series_episodes"`
	Status      animeStatus `xml:"series_status"`
	SeriesStart string      `xml:"series_start"`
	SeriesEnd   string      `xml:"series_end"`
	ImageURL    string      `xml:"series_image"`

	MyID                int        `xml:"my_id"`
	WatchedEpisodes     int        `xml:"my_watched_episodes"`
	MyStart             string     `xml:"my_start_date"`
	MyFinish            string     `xml:"my_finish_date"`
	MyScore             AnimeScore `xml:"my_score"`
	MyStatus            MyStatus   `xml:"my_status"`
	MyRewatching        int        `xml:"my_rewatching"`
	MyRewatchingEpisode int        `xml:"my_rewatching_ep"`
	LastUpdated         int        `xml:"my_last_updated"`
	MyTags              string     `xml:"my_tags"`
}

const AnimeXMLTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<entry>
	<episode>{{.WatchedEpisodes}}</episode>
	<status>{{.MyStatus}}</status>
	<score>{{.MyScore}}</score>
	<times_rewatched>{{.MyRewatching}}</times_rewatched>
	<rewatch_value>{{.MyRewatchingEpisode}}</rewatch_value>
	<date_start>{{.MyStart}}</date_start>
	<date_finish>{{.MyFinish}}</date_finish>
	<tags>{{.MyTags}}</tags>
</entry>`

type AnimeCustomSort struct {
	List  []*Anime
	LessF func(x, y *Anime) bool
}

func (acs AnimeCustomSort) Len() int {
	return len(acs.List)
}

func (acs AnimeCustomSort) Less(i, j int) bool {
	return acs.LessF(acs.List[i], acs.List[j])
}

func (acs AnimeCustomSort) Swap(i, j int) {
	acs.List[i], acs.List[j] = acs.List[j], acs.List[i]
}

func AnimeSortByLastUpdated(list []*Anime) sort.Interface {
	return AnimeCustomSort{list, func(x, y *Anime) bool {
		return x.LastUpdated > y.LastUpdated
	}}
}
