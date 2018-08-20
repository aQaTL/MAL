package anilist

var queryUserAnimeList = `
query UserList ($userID: Int) {
	MediaListCollection (userId: $userID, type: ANIME) {
		lists {
			entries {
				id
				status
				score(format: POINT_10)
				progress
				repeat
				updatedAt
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
					season
					episodes
					duration
					synonyms
				}
			}
			name
			isCustomList
			isSplitCompletedList
			status
		}
	}
}
`

var queryUserAnimeListFullDetails = `
query UserList ($userID: Int) {
	MediaListCollection (userId: $userID, type: ANIME) {
		lists {
			entries {
				id
				status
				score(format: POINT_10)
				progress
				updatedAt
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
					season
					episodes
					duration
					synonyms
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

var saveMediaListEntry = `
mutation ($listId: Int, $mediaId: Int, $status: MediaListStatus, $progress: Int, $score: Float) {
	SaveMediaListEntry (id: $listId, mediaId: $mediaId, status: $status, progress: $progress, score: $score) {
		id
		status
		progress
		score
		updatedAt
	}
}
`

var queryAiringSchedule = `
query ($mediaId: Int, $episode: Int) {
	AiringSchedule(mediaId: $mediaId, episode: $episode) {
		id
		airingAt
		timeUntilAiring
		episode
		mediaId
	}
}
`

var queryAiringNotification = `
query ($resetNotificationCount: Boolean) {
	Notification(type: AIRING, resetNotificationCount: $resetNotificationCount) {
		... on AiringNotification {
			id
			type
			animeId
			episode
			contexts
			createdAt
			media {
				title {
					romaji
					english
					native
					userPreferred				
				}
			}
		}
	}
}`