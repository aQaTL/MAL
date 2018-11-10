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
					` + mediaFull + `
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

var queryAiringNotifications = `
query ($page: Int, $perPage: Int, $resetNotificationCount: Boolean) {
	Page(page: $page, perPage: $perPage) {
		notifications(type_in: [AIRING], resetNotificationCount: $resetNotificationCount) {
			... on AiringNotification {
				id
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
	}
}
`

var deleteMediaListEntry = `
mutation ($id: Int) {
	DeleteMediaListEntry(id: $id) {
		deleted
	}
}
`

var mediaFull = `
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
episodes
duration
chapters
volumes
countryOfOrigin
isLicensed
source
hashtag
trailer {
	id
	site
}
updatedAt
coverImage {
	large
	medium
}
bannerImage
genres
synonyms
averageScore
meanScore
popularity
trending
tags {
	id
	name
	description
	category
	rank
	isGeneralSpoiler
	isMediaSpoiler
	isAdult
}
isFavourite
isAdult
nextAiringEpisode {
	id
	airingAt
	timeUntilAiring
	episode
}
siteUrl
`

var queryMedia = `
query ($page: Int, $perPage: Int, $search: String, $type: MediaType) {
	Page(page: $page, perPage: $perPage) {
		media(search: $search, type: $type) {
			` + mediaFull + `		
		}
	}
}
fragment FuzzyDateFields on FuzzyDate {
	year
	month
	day
}
`
