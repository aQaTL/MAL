package anilist

var queryUserAnimeList = `
query UserList ($userID: Int) {
	MediaListCollection (userId: $userID, type: ANIME) {
		lists {
			entries {
				id
				mediaId
				status
				score(format: POINT_10)
				progress
				media {
					id
					idMal
					title {
						romaji
						english
					}
					type
					format
					status
					description
					startDate {
						...FuzzyDateFields
					}
					endDate {
						...FuzzyDateFields
					}
					season
					duration
					source
					updatedAt
					genres
					synonyms
					averageScore
					meanScore
					popularity
					trending
					nextAiringEpisode {
						airingAt
						episode
					}
				}
			}
			name
			isCustomList
			isSplitCompletedList
			status
		}
	}
}
fragment FuzzyDateFields on FuzzyDate {
	year
	month
	day
}
`
