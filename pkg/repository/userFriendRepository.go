package repository

import (
	"context"
	"errors"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserFriendRepository struct{}

type UserFriendRequest struct {
	UserId primitive.ObjectID `json:"user_id" binding:"required"`
}

var (
	USER_RECOMMENDED = 0
	USER_PENDING     = 1
	USER_ACCEPTED    = 2
	USER_CANCELLED   = 3
)

// 1: Receiver 2: Sender

func (uf UserFriendRepository) Retrieve(c *gin.Context) {
	userType := c.Param("type")
	//userFriendsType, userFriendsTypeExists := c.Get("userfriends_type")
	userDetail := helpers.GetAuthUser(c)
	var UserFriends []models.UserFriends

	// Define pipeline for aggregation
	userFriendStatus := USER_ACCEPTED
	if userType != "" {
		userFriendStatus, _ = strconv.Atoi(userType)
	}

	userId := userDetail.UsersId
	uf.GetUser(c, userFriendStatus, userId, &UserFriends)

	if len(UserFriends) > 0 {
		for k := range UserFriends {
			uf.GetDetail(c, userDetail.UsersId, &UserFriends[k])
		}
		c.JSON(200, UserFriends)
	} else {
		helpers.ResponseNoData(c, "No data")
	}
}

func (uf UserFriendRepository) RetrieveOther(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var UserFriends []models.UserFriends
	uf.GetUser(c, USER_ACCEPTED, id, &UserFriends)

	if len(UserFriends) > 0 {
		for k := range UserFriends {
			uf.GetDetail(c, id, &UserFriends[k])
		}
		c.JSON(200, UserFriends)
	} else {
		helpers.ResponseNoData(c, "No data")
	}
}

func (uf UserFriendRepository) GetUser(c *gin.Context, userFriendStatus int, userId primitive.ObjectID, UserFriends *[]models.UserFriends) {
	userFriendsType, userFriendsTypeExists := c.Get("userfriends_type")
	userName := c.Query("name")
	userUsername := c.Query("username")

	filterStatus := bson.E{Key: "user_friends_status", Value: userFriendStatus}
	if userFriendStatus == USER_RECOMMENDED {
		filterStatus = bson.E{Key: "user_friends_status", Value: bson.M{"$in": []int{userFriendStatus, USER_PENDING}}}
	}

	match := bson.D{
		filterStatus,
		{
			Key: "$or", Value: []bson.D{
				{{Key: "user_friends_user_1", Value: userId}},
				{{Key: "user_friends_user_2", Value: userId}},
			},
		},
	}

	filter := bson.D{
		{Key: "$match", Value: match},
	}

	lookupUser1 := bson.D{{
		Key: "$lookup", Value: bson.M{
			"from":         "Users",
			"localField":   "user_friends_user_1",
			"foreignField": "_id",
			"as":           "User1",
		},
	}}

	lookupUser2 := bson.D{{
		Key: "$lookup", Value: bson.M{
			"from":         "Users",
			"localField":   "user_friends_user_2",
			"foreignField": "_id",
			"as":           "User2",
		},
	}}

	unwindLooupUser1 := bson.D{{
		Key: "$unwind", Value: "$User1",
	}}

	unwindLooupUser2 := bson.D{{
		Key: "$unwind", Value: "$User2",
	}}

	pipeline := mongo.Pipeline{
		filter,
	}

	if userName != "" || userUsername != "" {
		pipeline = append(pipeline, lookupUser1, lookupUser2, unwindLooupUser1, unwindLooupUser2)

		if userName != "" {
			matchUsername := bson.D{
				{
					Key: "$match", Value: bson.D{
						{Key: "$or", Value: []bson.M{
							{
								"User1._id":        bson.M{"$ne": userId},
								"User1.users_name": bson.M{"$regex": userName, "$options": "i"},
							},
							{
								"User2._id":        bson.M{"$ne": userId},
								"User2.users_name": bson.M{"$regex": userName, "$options": "i"},
							},
						}},
					},
				},
			}
			pipeline = append(pipeline, matchUsername)
		}

		if userUsername != "" {
			matchUserUsername := bson.D{
				{
					Key: "$match", Value: bson.D{
						{Key: "$or", Value: []bson.M{
							{
								"User1._id":            bson.M{"$ne": userId},
								"User1.users_username": bson.M{"$regex": userUsername, "$options": "i"},
							},
							{
								"User2._id":            bson.M{"$ne": userId},
								"User2.users_username": bson.M{"$regex": userUsername, "$options": "i"},
							},
						}},
					},
				},
			}
			pipeline = append(pipeline, matchUserUsername)
		}
	}

	if userFriendStatus == 0 {
		sortBy := bson.D{{
			Key: "$sort", Value: bson.M{"user_friends_is_official": -1, "user_friends_created_at": 1},
		}}
		pipeline = append(pipeline, sortBy)
	}

	cursor, _ := config.DB.Collection("UserFriends").Aggregate(context.TODO(), pipeline)
	cursor.All(context.TODO(), UserFriends)

	if userFriendsTypeExists {
		if userFriendsType == strconv.Itoa(USER_RECOMMENDED) {
			var UserCurrentFriends []models.UserFriends
			var UserFriendsId []primitive.ObjectID

			filterFriends := bson.D{
				{
					Key: "$or", Value: []bson.D{
						{{Key: "user_friends_user_1", Value: userId}},
						{{Key: "user_friends_user_2", Value: userId}},
					},
				},
			}
			cursorUserFriends, _ := config.DB.Collection("UserFriends").Find(context.TODO(), filterFriends)
			cursorUserFriends.All(context.TODO(), &UserCurrentFriends)

			UserFriendsId = append(UserFriendsId, userId)
			for _, v := range UserCurrentFriends {
				if v.UserFriendsUser1 == userId {
					UserFriendsId = append(UserFriendsId, v.UserFriendsUser2)
				} else {
					UserFriendsId = append(UserFriendsId, v.UserFriendsUser1)
				}
			}

			var UserNotFriends []models.Users
			filterNotFriends := bson.D{{Key: "_id", Value: bson.M{"$nin": UserFriendsId}}}

			if userName != "" {
				filterNotFriends = append(filterNotFriends, bson.E{Key: "users_name", Value: bson.M{"$regex": userName, "$options": "i"}})
			}

			if userUsername != "" {
				filterNotFriends = append(filterNotFriends, bson.E{Key: "users_username", Value: bson.M{"$regex": userName, "$options": "i"}})
			}

			cursorNotFriends, _ := config.DB.Collection("Users").Find(context.TODO(), filterNotFriends)
			cursorNotFriends.All(context.TODO(), &UserNotFriends)

			isOfficial := false
			for _, v := range UserNotFriends {
				*UserFriends = append(*UserFriends, models.UserFriends{
					UserFriendsStatus:     &USER_RECOMMENDED,
					UserFriendsUser1:      userId,
					UserFriendsUser2:      v.UsersId,
					UserFriendsIsOfficial: &isOfficial,
				})
			}
		}
	}
}

func (uf UserFriendRepository) Create(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)

	var UserAddedDetail models.Users
	var payload UserFriendRequest
	validateError := helpers.Validate(c, &payload)
	if validateError != nil {
		return
	}

	if payload.UserId == userDetail.UsersId {
		helpers.ResponseBadRequestError(c, "Cannot send request to yourself")
		return
	}

	userAddedDetailFilter := bson.D{{Key: "_id", Value: payload.UserId}}
	userAddedDetailErr := config.DB.Collection("Users").FindOne(context.TODO(), userAddedDetailFilter).Decode(&UserAddedDetail)

	if userAddedDetailErr == mongo.ErrNoDocuments {
		helpers.ResponseBadRequestError(c, "Invalid user")
		return
	}

	exceedError := uf.CheckIfExceedLimit(userDetail)
	if exceedError != nil {
		helpers.ResponseBadRequestError(c, "Your "+exceedError.Error())
		return
	}

	exceedSentError := uf.CheckIfExceedLimit(UserAddedDetail)
	if exceedSentError != nil {
		helpers.ResponseBadRequestError(c, "Receiving user's "+exceedSentError.Error())
		return
	}

	status := USER_PENDING
	if UserAddedDetail.UsersSettingFriendAutoAdd != nil && *UserAddedDetail.UsersSettingFriendAutoAdd == 1 {
		status = USER_ACCEPTED
	}

	ins := models.UserFriends{
		UserFriendsStatus:    &status,
		UserFriendsUser1:     payload.UserId,
		UserFriendsUser2:     userDetail.UsersId,
		UserFriendsCreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	var UserFriends models.UserFriends
	err := uf.CheckIfFriend(c, userDetail, userDetail.UsersId, payload.UserId, &UserFriends)
	isNewRecord := false

	if err == mongo.ErrNoDocuments {
		isNewRecord = true
	}

	if isNewRecord {
		var NewUserFriends models.UserFriends
		result, _ := config.DB.Collection("UserFriends").InsertOne(context.TODO(), ins)
		config.DB.Collection("UserFriends").FindOne(context.TODO(), bson.D{{Key: "_id", Value: result.InsertedID}}).Decode(&NewUserFriends)
		uf.GetDetail(c, userDetail.UsersId, &NewUserFriends)

		if status == USER_PENDING {
			uf.HandleNotificationsPending(c, userDetail, UserAddedDetail, NewUserFriends)
		} else if status == USER_ACCEPTED {
			uf.CountFriends(NewUserFriends.UserFriendsUser1)
			uf.CountFriends(NewUserFriends.UserFriendsUser2)
			uf.HandleNotificationsAccepted(c, userDetail, UserAddedDetail, NewUserFriends)
		}

		uf.GetDetail(c, userDetail.UsersId, &UserFriends)
		c.JSON(200, NewUserFriends)
	} else {
		if UserFriends.UserFriendsUser2 == userDetail.UsersId {
			if *UserFriends.UserFriendsStatus == USER_PENDING {
				helpers.ResponseBadRequestError(c, "Unable to add a user that is pending confirmation")
				return
			} else if *UserFriends.UserFriendsStatus == USER_ACCEPTED {
				helpers.ResponseBadRequestError(c, "Unable to add a user that is in your friend list")
				return
			} else if *UserFriends.UserFriendsStatus == USER_RECOMMENDED {
				isAdded := false
				*UserFriends.UserFriendsStatus = USER_PENDING
				if *UserAddedDetail.UsersSettingFriendAutoAdd == 1 {
					*UserFriends.UserFriendsStatus = USER_ACCEPTED
					isAdded = true
				}

				filter := bson.D{{Key: "_id", Value: UserFriends.UserFriendsId}}
				update := bson.D{{Key: "$set", Value: UserFriends}}
				config.DB.Collection("UserFriends").UpdateOne(context.TODO(), filter, update)

				if isAdded {
					uf.CountFriends(UserFriends.UserFriendsUser1)
					uf.CountFriends(UserFriends.UserFriendsUser2)
					uf.HandleNotificationsAccepted(c, userDetail, UserAddedDetail, UserFriends)
				} else {
					uf.HandleNotificationsPending(c, userDetail, UserAddedDetail, UserFriends)
				}
			}
		} else if UserFriends.UserFriendsUser1 == userDetail.UsersId {
			if *UserFriends.UserFriendsStatus == USER_PENDING {
				*UserFriends.UserFriendsStatus = USER_ACCEPTED
				filter := bson.D{{Key: "_id", Value: UserFriends.UserFriendsId}}
				update := bson.D{{Key: "$set", Value: UserFriends}}
				config.DB.Collection("UserFriends").UpdateOne(context.TODO(), filter, update)

				uf.CountFriends(UserFriends.UserFriendsUser1)
				uf.CountFriends(UserFriends.UserFriendsUser2)
			} else if *UserFriends.UserFriendsStatus == USER_RECOMMENDED {
				isAdded := false
				*UserFriends.UserFriendsStatus = USER_PENDING
				if *UserAddedDetail.UsersSettingFriendAutoAdd == 1 {
					*UserFriends.UserFriendsStatus = USER_ACCEPTED
					isAdded = true
				}

				// Flip receivre as user2
				UserFriends.UserFriendsUser1 = UserFriends.UserFriendsUser2
				UserFriends.UserFriendsUser2 = userDetail.UsersId

				filter := bson.D{{Key: "_id", Value: UserFriends.UserFriendsId}}
				update := bson.D{{Key: "$set", Value: UserFriends}}
				config.DB.Collection("UserFriends").UpdateOne(context.TODO(), filter, update)

				if isAdded {
					uf.CountFriends(UserFriends.UserFriendsUser1)
					uf.CountFriends(UserFriends.UserFriendsUser2)
					uf.HandleNotificationsAccepted(c, userDetail, UserAddedDetail, UserFriends)
				} else {
					uf.HandleNotificationsPending(c, userDetail, UserAddedDetail, UserFriends)
				}

			}
		}
		uf.GetDetail(c, userDetail.UsersId, &UserFriends)
		c.JSON(200, UserFriends)
	}
}

func (uf UserFriendRepository) CheckIfFriend(c *gin.Context, userDetail models.Users, user1 primitive.ObjectID, user2 primitive.ObjectID, UserFriends *models.UserFriends) error {
	err := config.DB.Collection("UserFriends").FindOne(context.TODO(), bson.D{{
		Key: "$or", Value: []bson.D{
			{
				{Key: "user_friends_user_1", Value: user2},
				{Key: "user_friends_user_2", Value: user1},
			},
			{
				{Key: "user_friends_user_1", Value: user1},
				{Key: "user_friends_user_2", Value: user2},
			},
		},
	}}).Decode(&UserFriends)

	uf.GetDetail(c, userDetail.UsersId, UserFriends)
	return err
}

func (uf UserFriendRepository) GetDetail(c *gin.Context, id primitive.ObjectID, UserFriends *models.UserFriends) {
	var friendId primitive.ObjectID
	if UserFriends.UserFriendsUser1 == id {
		friendId = UserFriends.UserFriendsUser2
	} else {
		friendId = UserFriends.UserFriendsUser1
	}

	var User models.UsersAggBreathing
	config.DB.Collection("Users").FindOne(context.TODO(), bson.D{
		{Key: "_id", Value: friendId},
	}).Decode(&User)

	User.UserBreathingStatus = AuthRepository{}.Breathing(c, User.UsersId)
	UserFriends.UserFriendsDetail = &User
}

func (uf UserFriendRepository) Update(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("userFriendId"))
	userDetail := helpers.GetAuthUser(c)

	var UserFriends models.UserFriends
	var UserRequest models.Users
	err := config.DB.Collection("UserFriends").FindOne(context.TODO(), bson.D{
		{Key: "_id", Value: id},
		{Key: "user_friends_user_1", Value: userDetail.UsersId},
		{Key: "user_friends_status", Value: USER_PENDING},
	}).Decode(&UserFriends)

	if err == mongo.ErrNoDocuments {
		helpers.ResponseBadRequestError(c, "no request to approve")
		return
	}

	exceedError := uf.CheckIfExceedLimit(userDetail)
	if exceedError != nil {
		helpers.ResponseBadRequestError(c, exceedError.Error())
		return
	}

	*UserFriends.UserFriendsStatus = USER_ACCEPTED
	filter := bson.D{{Key: "_id", Value: UserFriends.UserFriendsId}}
	update := bson.D{{Key: "$set", Value: UserFriends}}
	config.DB.Collection("UserFriends").UpdateOne(context.TODO(), filter, update)

	filterRequester := bson.D{{Key: "_id", Value: UserFriends.UserFriendsUser2}}
	config.DB.Collection("Users").FindOne(context.TODO(), filterRequester).Decode(&UserRequest)

	uf.CountFriends(UserFriends.UserFriendsUser1)
	uf.CountFriends(UserFriends.UserFriendsUser2)
	uf.HandleNotificationsAccepted(c, userDetail, UserRequest, UserFriends)

	helpers.ResponseSuccessMessage(c, "Friend request accepted")
}

func (uf UserFriendRepository) Delete(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("userFriendId"))
	userDetail := helpers.GetAuthUser(c)

	var UserFriends models.UserFriends
	filterCheck := bson.D{
		{Key: "_id", Value: id},
		{
			Key: "$or", Value: []bson.D{
				{{Key: "user_friends_user_1", Value: userDetail.UsersId}},
				{{Key: "user_friends_user_2", Value: userDetail.UsersId}},
			},
		},
		{
			Key: "$or", Value: []bson.D{
				{{Key: "user_friends_status", Value: bson.D{{Key: "$ne", Value: USER_PENDING}}}},
				{{Key: "user_friends_status", Value: bson.D{{Key: "$ne", Value: USER_RECOMMENDED}}}},
			},
		},
	}
	err := config.DB.Collection("UserFriends").FindOne(context.TODO(), filterCheck).Decode(&UserFriends)

	if err == mongo.ErrNoDocuments {
		helpers.ResponseBadRequestError(c, "no request to reject")
		return
	}

	if UserFriends.UserFriendsUser1 != userDetail.UsersId && *UserFriends.UserFriendsStatus == USER_PENDING {
		//
		helpers.ResponseBadRequestError(c, "Unable to reject. You are the friend requester")
		return
	}

	*UserFriends.UserFriendsStatus = USER_CANCELLED
	filter := bson.D{{Key: "_id", Value: UserFriends.UserFriendsId}}
	update := bson.D{{Key: "$set", Value: UserFriends}}
	config.DB.Collection("UserFriends").UpdateOne(context.TODO(), filter, update)

	uf.CountFriends(UserFriends.UserFriendsUser1)
	uf.CountFriends(UserFriends.UserFriendsUser2)

	helpers.ResponseSuccessMessage(c, "Friend request deleted")
}

func (uf UserFriendRepository) Recommended(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)
	var officialAccount []models.Users
	officialAccountFilter := bson.D{
		{Key: "$match", Value: bson.M{
			"users_is_business": bson.M{"$exists": true},
			"_id":               bson.M{"$ne": userDetail.UsersId},
		}},
	}

	lookupUser1 := bson.D{{
		Key: "$lookup", Value: bson.M{
			"as":   "UserFriends1",
			"from": "UserFriends",
			"let":  bson.M{"user_id": "$_id"},
			"pipeline": []bson.M{
				{"$match": bson.M{
					"$expr": bson.M{
						"$eq": bson.A{"$user_friends_user_1", "$$user_id"},
					},
				}},
				{"$match": bson.M{"user_friends_user_2": userDetail.UsersId}},
			},
		},
	}}

	lookupUser2 := bson.D{{
		Key: "$lookup", Value: bson.M{
			"as":   "UserFriends2",
			"from": "UserFriends",
			"let":  bson.M{"user_id": "$_id"},
			"pipeline": []bson.M{
				{"$match": bson.M{
					"$expr": bson.M{
						"$eq": bson.A{"$user_friends_user_2", "$$user_id"},
					},
				}},
				{"$match": bson.M{"user_friends_user_1": userDetail.UsersId}},
			},
		},
	}}

	unwindLooupUser1 := bson.D{{
		Key: "$unwind", Value: bson.M{"path": "$UserFriends1", "preserveNullAndEmptyArrays": true},
	}}

	unwindLooupUser2 := bson.D{{
		Key: "$unwind", Value: bson.M{"path": "$UserFriends2", "preserveNullAndEmptyArrays": true},
	}}

	excludeFriend := bson.D{{
		Key: "$match", Value: bson.M{
			"UserFriends1._id": bson.M{"$exists": false},
			"UserFriends2._id": bson.M{"$exists": false},
		},
	}}

	officialAccountPipeline := mongo.Pipeline{
		officialAccountFilter,
		lookupUser1,
		lookupUser2,
		unwindLooupUser1,
		unwindLooupUser2,
		excludeFriend,
	}

	cursor, _ := config.DB.Collection("Users").Aggregate(context.TODO(), officialAccountPipeline)
	cursor.All(context.TODO(), &officialAccount)

	var ins []interface{}

	isOfficial := true
	for _, v := range officialAccount {
		ins = append(ins, models.UserFriends{
			UserFriendsStatus:     &USER_RECOMMENDED,
			UserFriendsUser1:      v.UsersId,
			UserFriendsUser2:      userDetail.UsersId,
			UserFriendsIsOfficial: &isOfficial,
			UserFriendsCreatedAt:  primitive.NewDateTimeFromTime(time.Now()),
		})
	}
	if len(ins) > 0 {
		config.DB.Collection("UserFriends").InsertMany(context.TODO(), ins)
	}

	c.Set("userfriends_type", strconv.Itoa(USER_RECOMMENDED))
	c.AddParam("type", strconv.Itoa(USER_RECOMMENDED))
	uf.Retrieve(c)
}

func (uf UserFriendRepository) CountFriends(userId primitive.ObjectID) {
	countFilter := bson.M{
		"$or": []bson.M{
			{"user_friends_user_1": userId},
			{"user_friends_user_2": userId},
		},
		"user_friends_status": USER_ACCEPTED,
	}

	friendCount, _ := config.DB.Collection("UserFriends").CountDocuments(context.TODO(), countFilter)
	update := bson.D{{Key: "$set", Value: bson.M{
		"users_friends_count": int(friendCount),
	}}}
	config.DB.Collection("Users").UpdateOne(context.TODO(), bson.D{{Key: "_id", Value: userId}}, update)
}

func (uf UserFriendRepository) CheckIfExceedLimit(userDetail models.Users) error {
	if userDetail.UsersFriendsCount+1 > int(config.APP_LIMIT.FriendListLimit) {
		errMessage := "friend list has reached limit of " + strconv.Itoa(userDetail.UsersFriendsCount) + "/" + strconv.Itoa(int(config.APP_LIMIT.FriendListLimit))
		return errors.New(errMessage)
	}
	return nil
}

func (uf UserFriendRepository) HandleNotificationsPending(c *gin.Context, UserDetail models.Users, User2Detail models.Users, UserFriends models.UserFriends) {
	NotificationMessage := models.NotificationMessage{
		Message: "{0}發送了好友邀請給你",
		Data:    []map[string]interface{}{helpers.NotificationFormatUser(UserDetail), helpers.NotificationFormatUserFriends(UserFriends)},
	}
	helpers.NotificationsCreate(c, helpers.NOTIFICATION_FRIEND_REQUEST, User2Detail.UsersId, NotificationMessage, UserFriends.UserFriendsId)
}

func (uf UserFriendRepository) HandleNotificationsAccepted(c *gin.Context, UserDetail models.Users, User2Detail models.Users, UserFriends models.UserFriends) {
	// Side 1
	NotificationMessage := models.NotificationMessage{
		Message: "{0}成為你的好友",
		Data:    []map[string]interface{}{helpers.NotificationFormatUser(UserDetail), helpers.NotificationFormatUserFriends(UserFriends)},
	}
	helpers.NotificationsCreate(c, helpers.NOTIFICATION_FRIEND_REQUEST_ACCEPTED, User2Detail.UsersId, NotificationMessage, UserFriends.UserFriendsId)

	NotificationMessage2 := models.NotificationMessage{
		Message: "{0}成為你的好友",
		Data:    []map[string]interface{}{helpers.NotificationFormatUser(User2Detail), helpers.NotificationFormatUserFriends(UserFriends)},
	}
	helpers.NotificationsCreate(c, helpers.NOTIFICATION_FRIEND_REQUEST_ACCEPTED, UserDetail.UsersId, NotificationMessage2, UserFriends.UserFriendsId)
}
