package helpers

import (
	"context"
	"errors"
	"oosa/internal/config"
	"oosa/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	NOTIFICATION_INVITATION              = "INVITATION"
	NOTIFICATION_JOINING_NEW             = "JOINING_NEW"
	NOTIFICATION_JOINING_REQUEST         = "JOINING_REQUEST"
	NOTIFICATION_FRIEND_REQUEST          = "FRIEND_REQUEST"
	NOTIFICATION_FRIEND_REQUEST_ACCEPTED = "FRIEND_REQUEST_ACCEPTED"
	NOTIFICATION_BADGE_NEW               = "BADGE_NEW"
	NOTIFICATION_STARMAP_NEW             = "STARMAP_NEW"
	NOTIFICATION_STARMAP_PROGRESS        = "STARMAP_PROGRESS"
	NOTIFICATION_UPDATE_POLICY           = "UPDATE_POLICY"
	NOTIFICATION_EVENT_INFO              = "EVENT_INFO"
	NOTIFICATION_EVENT_COUNTDOWN         = "EVENT_COUNTDOWN"
	NOTIFICATION_EVENT_MEMBER_DELETED    = "EVENT_MEMBER_DELETED"
	NOTIFICATION_EVENT_JOIN_DENIED       = "EVENT_JOIN_DENIED"
	NOTIFICATION_COLOG_PHOTO_UPLOADED    = "COLOG_PHOTO_UPLOADED"
	NOTIFICATION_COLOG_REMIND            = "COLOG_REMIND"
)

const (
	NOTIFICATION_STATE_NEW  = "new"
	NOTIFICATION_STATE_DONE = "done"
)

func NotificationsCreate(c *gin.Context, notifCode string, userId primitive.ObjectID, message models.NotificationMessage, identifier primitive.ObjectID) {
	userDetail := GetAuthUser(c)
	insert := models.Notifications{
		NotificationsCode:       notifCode,
		NotificationsUser:       userId,
		NotificationsMessage:    message,
		NotificationsIdentifier: identifier,
		NotificationsCreatedAt:  primitive.NewDateTimeFromTime(time.Now()),
		NotificationsCreatedBy:  userDetail.UsersId,
		NotificationsState:      NOTIFICATION_STATE_NEW,
	}
	config.DB.Collection("Notifications").InsertOne(context.TODO(), insert)
}

const (
	header_notificationId = "X-Notification-Id"
)

func IsHeaderHasNotifcationId(c *gin.Context) bool {
	header := c.GetHeader(header_notificationId)
	if header == "" {
		return false
	}
	_, err := primitive.ObjectIDFromHex(header)
	return err == nil
}

func NotificationsUpdateState(c *gin.Context, state string) error {
	header := c.GetHeader(header_notificationId)
	if header == "" {
		return errors.New("header X-Notification-Id is required")
	}

	notificationID, err := primitive.ObjectIDFromHex(header)
	if err != nil {
		return errors.New("invalid notification ID")
	}
	filter := bson.D{
		{Key: "_id", Value: notificationID},
	}
	update := bson.D{
		{Key: "$set", Value: bson.M{
			"notifications_state": state,
		}},
	}
	config.DB.Collection("Notifications").UpdateOne(context.TODO(), filter, update)
	return nil
}

func NotificationFormatEvent(Events models.Events) map[string]any {
	return map[string]any{
		"events_id":   Events.EventsId,
		"events_name": Events.EventsName,
	}
}

func NotificationFormatUser(Users models.Users) map[string]any {
	return map[string]any{
		"users_id":   Users.UsersId,
		"users_name": Users.UsersName,
	}
}

func NotificationFormatUserFollowing(UserFollowings models.UserFollowings) map[string]any {
	return map[string]any{
		"user_followings_id":        UserFollowings.UserFollowingsId,
		"user_followings_user":      UserFollowings.UserFollowingsUser,
		"user_followings_following": UserFollowings.UserFollowingsFollowing,
	}
}

func NotificationFormatBadges(Badges models.Badges) map[string]any {
	return map[string]any{
		"badges_id":   Badges.BadgesId,
		"badges_code": Badges.BadgesCode,
		"badges_name": Badges.BadgesName,
	}
}

func NotificationFormatUserFriends(UserFriends models.UserFriends) map[string]any {
	return map[string]any{
		"user_friends_id":     UserFriends.UserFriendsId,
		"user_friends_user_1": UserFriends.UserFriendsUser1,
		"user_friends_user_2": UserFriends.UserFriendsUser2,
	}
}
