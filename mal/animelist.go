package mal

type AnimeList []*Anime

func (list AnimeList) GetByID(id int) *Anime {
	for _, entry := range list {
		if entry.ID == id {
			return entry
		}
	}
	return nil
}

func (list AnimeList) FilterByStatus(status MyStatus) AnimeList {
	filter := func(vs AnimeList, f func(anime *Anime) bool) AnimeList {
		vsf := make(AnimeList, 0)
		for _, a := range vs {
			if f(a) {
				vsf = append(vsf, a)
			}
		}
		return vsf
	}

	if status == All {
		return list
	} else {
		return filter(list, func(anime *Anime) bool { return anime.MyStatus == status })
	}
}
