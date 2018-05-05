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

func (list AnimeList) DeleteByID(id int) AnimeList {
	idx := 0
	found := false
	for i, entry := range list {
		if entry.ID == id {
			idx = i
			found = true
			break
		}
	}
	if !found {
		return list
	}

	copy(list[idx:], list[idx+1:])
	list[len(list)-1] = nil
	list = list[:len(list)-1]

	return list
}