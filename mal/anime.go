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

type animeScore int

const (
	NotRatedYet animeScore = iota
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

type myStatus int

const (
	All         myStatus = iota
	Watching
	Completed
	OnHold
	Dropped
	PlanToWatch myStatus = 6
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
	MyScore             animeScore `xml:"my_score"`
	MyStatus            myStatus   `xml:"my_status"`
	MyRewatching        int        `xml:"my_rewatching"`
	MyRewatchingEpisode int        `xml:"my_rewatching_ep"`
	LastUpdated         int        `xml:"my_last_updated"`
	MyTags              string     `xml:"my_tags"`
}

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