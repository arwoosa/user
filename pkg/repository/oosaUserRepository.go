package repository

import (
	"context"
	"net/http"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OosaUserRepository struct{}

func (r OosaUserRepository) Read(c *gin.Context) {
	userIdVal := c.Param("id")
	userId, _ := primitive.ObjectIDFromHex(userIdVal)
	userDetail := helpers.GetAuthUser(c)

	var User models.Users
	err := r.ReadUserById(userId, &User)

	if err == mongo.ErrNoDocuments {
		helpers.ResponseNoData(c, "")
		return
	}

	// 添加追蹤關係
	var UserFollowing models.UserFollowings
	UserFollowingFilter := bson.D{
		{Key: "user_followings_user", Value: userDetail.UsersId},
		{Key: "user_followings_following", Value: User.UsersId},
	}
	UserFollowingErr := config.DB.Collection("UserFollowings").FindOne(context.TODO(), UserFollowingFilter).Decode(&UserFollowing)
	if UserFollowingErr == nil {
		User.UsersFollowings = &UserFollowing
	}

	// 添加好友關係信息
	userFriendStatus := "stranger"
	if userDetail.UsersId == User.UsersId {
		userFriendStatus = "myself"
	} else {
		var UserFriends models.UserFriends
		UserFriendsFilter := bson.D{
			{Key: "$or", Value: []bson.D{
				{
					{Key: "user_friends_user_1", Value: userDetail.UsersId},
					{Key: "user_friends_user_2", Value: User.UsersId},
				},
				{
					{Key: "user_friends_user_1", Value: User.UsersId},
					{Key: "user_friends_user_2", Value: userDetail.UsersId},
				},
			}},
		}

		UserFriendsErr := config.DB.Collection("UserFriends").FindOne(context.TODO(), UserFriendsFilter).Decode(&UserFriends)
		if UserFriendsErr == nil {
			// 將狀態碼轉換為文字狀態
			// USER_RECOMMENDED = 0, USER_PENDING = 1, USER_ACCEPTED = 2, USER_CANCELLED = 3
			switch *UserFriends.UserFriendsStatus {
			case 2: // USER_ACCEPTED
				userFriendStatus = "friend"
			case 1: // USER_PENDING
				// 判斷是誰發出的邀請
				if (UserFriends.UserFriendsUser1 == userDetail.UsersId && UserFriends.UserFriendsUser2 == User.UsersId) ||
					(UserFriends.UserFriendsUser2 == userDetail.UsersId && UserFriends.UserFriendsUser1 == User.UsersId) {
					userFriendStatus = "invited"
				}
			default:
				userFriendStatus = "stranger"
			}
		}
	}

	// 獲取好友數量
	friendsCount := User.UsersFriendsCount

	// 添加到響應中
	usersFriendsInfo := map[string]interface{}{
		"status": userFriendStatus,
		"count":  friendsCount,
	}

	// 獲取用戶的 breathing points
	breathingPoints := AuthRepository{}.Breathing(c, User.UsersId)

	// 創建最終響應
	response := map[string]interface{}{
		"users_avatar":                              User.UsersAvatar,
		"users_created_at":                          User.UsersCreatedAt,
		"users_email":                               User.UsersEmail,
		"users_id":                                  User.UsersId,
		"users_is_business":                         User.UsersIsBusiness,
		"users_is_subscribed":                       User.UsersIsSubscribed,
		"users_name":                                User.UsersName,
		"users_object":                              User.UsersObject,
		"users_password":                            User.UsersPassword,
		"users_setting_friend_auto_add":             User.UsersSettingFriendAutoAdd,
		"users_setting_is_visible_friends":          User.UsersSettingIsVisibleFriends,
		"users_setting_is_visible_statistics":       User.UsersSettingIsVisibleStatistics,
		"users_setting_language":                    User.UsersSettingLanguage,
		"users_setting_visibility_activity_summary": User.UsersSettingVisibilityActivitySummary,
		"users_source":                              User.UsersSource,
		"users_source_id":                           User.UsersSourceId,
		"users_username":                            User.UsersUsername,
		"users_breathing_points":                    breathingPoints,
		"users_following_count":                     User.UsersFollowingCount,
		"users_follower_count":                      User.UsersFollowerCount,
		"users_friends_info":                        usersFriendsInfo,
	}

	// 如果有追蹤關係，添加到響應中
	if User.UsersFollowings != nil {
		response["users_followings"] = User.UsersFollowings
	}

	c.JSON(http.StatusOK, response)
}

func (r OosaUserRepository) ReadUserById(userId primitive.ObjectID, User *models.Users) error {
	err := config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "_id", Value: userId}}).Decode(&User)
	return err
}
