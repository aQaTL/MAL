package anilist

var queryUserAnimeList = `
query UserList ($userID: Int) {
	MediaListCollection (userId: $userID, type: ANIME) {
		lists {
			entries {
				status
				score(format: POINT_10)
				progress
				media {
					id
					idMal
					title {
						romaji
						english
						native
						userPreferred
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
					popularity
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

var queryAuthenticatedUser = `
query {
	Viewer {
		id
		name
		about
		bannerImage
		stats {
			watchedTime
		}
		unreadNotificationCount
		siteUrl
		donatorTier
		moderatorStatus
		updatedAt
	}
}
`