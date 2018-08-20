package anilist

import (
	"github.com/aqatl/mal/oauth2"
	"github.com/aqatl/cliwait"
)

func QueryUserListsWaitAnimation(userId int, token oauth2.OAuthToken) ([]MediaListGroup, error) {
	var mlg []MediaListGroup
	var err error
	cliwait.DoFuncWithWaitAnimation("Queyring user list", func() {
		mlg, err = QueryUserLists(userId, token)
	})
	return mlg, err
}

func QueryAuthenticatedUserWaitAnimation(user *User, token oauth2.OAuthToken) error {
	var err error
	cliwait.DoFuncWithWaitAnimation("Saving entry", func() {
		err = QueryAuthenticatedUser(user, token)
	})
	return err
}

func SaveMediaListEntryWaitAnimation(entry *MediaListEntry, token oauth2.OAuthToken) error {
	var err error
	cliwait.DoFuncWithWaitAnimation("Saving entry", func() {
		err = SaveMediaListEntry(entry, token)
	})
	return err
}

func QueryAiringScheduleWaitAnimation(mediaId, episode int, token oauth2.OAuthToken) (
	AiringSchedule, error,
) {
	var as AiringSchedule
	var err error
	cliwait.DoFuncWithWaitAnimation("Querying airing schedule", func() {
		as, err = QueryAiringSchedule(mediaId, episode, token)
	})
	return as, err
}

func QueryAiringNotificationsWaitAnimation(page, perPage int, markRead bool, token oauth2.OAuthToken) (
	[]AiringNotification, error,
) {
	var n []AiringNotification
	var err error
	cliwait.DoFuncWithWaitAnimation("Querying notification", func() {
		n, err = QueryAiringNotifications(page, perPage, markRead, token)
	})
	return n, err
}
