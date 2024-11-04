package repository

import (
	"context"
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

//0: recommended, 1: pending, 2: accepted, 3: cancel

func (uf UserFriendRepository) Retrieve(c *gin.Context) {
	userType := c.Param("type")
	//userFriendsType, userFriendsTypeExists := c.Get("userfriends_type")
	userDetail := helpers.GetAuthUser(c)
	var UserFriends []models.UserFriends

	// Define pipeline for aggregation
	userFriendStatus := 2
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
	uf.GetUser(c, 2, id, &UserFriends)

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
	userName := c.Query("name")
	match := bson.D{
		{Key: "user_friends_status", Value: userFriendStatus},
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
		pipeline = append(pipeline, lookupUser1, lookupUser2, unwindLooupUser1, unwindLooupUser2, matchUsername)
	}

	cursor, _ := config.DB.Collection("UserFriends").Aggregate(context.TODO(), pipeline)
	cursor.All(context.TODO(), UserFriends)
}

func (uf UserFriendRepository) Create(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)

	var payload UserFriendRequest
	validateError := helpers.Validate(c, &payload)
	if validateError != nil {
		return
	}

	if payload.UserId == userDetail.UsersId {
		helpers.ResponseBadRequestError(c, "Cannot send request to yourself")
		return
	}

	ins := models.UserFriends{
		UserFriendsStatus:    1,
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
		c.JSON(200, NewUserFriends)
	} else {
		if UserFriends.UserFriendsUser2 == userDetail.UsersId {
			if UserFriends.UserFriendsStatus == 1 {
				helpers.ResponseBadRequestError(c, "Unable to add a user that is pending confirmation")
				return
			} else if UserFriends.UserFriendsStatus == 2 {
				helpers.ResponseBadRequestError(c, "Unable to add a user that is in your friend list")
				return
			}
		} else if UserFriends.UserFriendsUser1 == userDetail.UsersId {
			UserFriends.UserFriendsStatus = 2
			filter := bson.D{{Key: "_id", Value: UserFriends.UserFriendsId}}
			update := bson.D{{Key: "$set", Value: UserFriends}}
			config.DB.Collection("UserFriends").UpdateOne(context.TODO(), filter, update)
		}
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
	err := config.DB.Collection("UserFriends").FindOne(context.TODO(), bson.D{
		{Key: "_id", Value: id},
		{Key: "user_friends_user_1", Value: userDetail.UsersId},
		{Key: "user_friends_status", Value: 1},
	}).Decode(&UserFriends)

	if err == mongo.ErrNoDocuments {
		helpers.ResponseBadRequestError(c, "no request to approve")
		return
	}

	UserFriends.UserFriendsStatus = 2
	filter := bson.D{{Key: "_id", Value: UserFriends.UserFriendsId}}
	update := bson.D{{Key: "$set", Value: UserFriends}}
	config.DB.Collection("UserFriends").UpdateOne(context.TODO(), filter, update)

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
				{{Key: "user_friends_status", Value: bson.D{{Key: "$ne", Value: 0}}}},
				{{Key: "user_friends_status", Value: bson.D{{Key: "$ne", Value: 3}}}},
			},
		},
	}
	err := config.DB.Collection("UserFriends").FindOne(context.TODO(), filterCheck).Decode(&UserFriends)

	if err == mongo.ErrNoDocuments {
		helpers.ResponseBadRequestError(c, "no request to reject")
		return
	}

	if UserFriends.UserFriendsUser1 != userDetail.UsersId && UserFriends.UserFriendsStatus == 1 {
		//
		helpers.ResponseBadRequestError(c, "Unable to reject. You are the friend requester")
		return
	}

	UserFriends.UserFriendsStatus = 3
	filter := bson.D{{Key: "_id", Value: UserFriends.UserFriendsId}}
	update := bson.D{{Key: "$set", Value: UserFriends}}
	config.DB.Collection("UserFriends").UpdateOne(context.TODO(), filter, update)

	helpers.ResponseSuccessMessage(c, "Friend request deleted")
}

func (uf UserFriendRepository) Recommended(c *gin.Context) {
	c.Set("userfriends_type", "0")
	c.AddParam("type", "0")
	uf.Retrieve(c)
}
