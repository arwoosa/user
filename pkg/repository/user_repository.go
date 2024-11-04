package repository

import (
	"context"
	"net/http"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct{}
type UserFollowingsRequest struct {
	// UserFollowingsUser      primitive.ObjectID `json:"user_followings_user" binding:"required"`
	UserFollowingsFollowing primitive.ObjectID `json:"user_followings_following" binding:"required"`
}

// @Summary UserFollowings
// @Description Retrieve all userfollowings
// @ID userfollowings
// @Produce json
// @Tags UserFollowings
// @Success 200 {object} []models.UserFollowings
// @Failure 400 {object} structs.Message
// @Router /userfollowings [get]
func (uf UserRepository) UserFollowingRetrieve(c *gin.Context) {
	var results []models.UserFollowings

	cursor, err := config.DB.Collection("UserFollowings").Find(context.TODO(), bson.D{})
	if err != nil {
		panic(err)
	}
	cursor.All(context.TODO(), &results)

	if len(results) == 0 {
		helpers.ResponseNoData(c, "No Data")
		return
	}
	c.JSON(200, results)
}

// @Summary UserFollowings
// @Description Create userfollowings data
// @ID userfollowings
// @Produce json
// @Tags UserFollowings
// @Success 200 {object} Message and models.UserFollowings
// @Failure 400 {object} structs.Message
// @Router /userfollowings [post]
func (uf UserRepository) UserFollowingCreate(c *gin.Context) {
	var payload UserFollowingsRequest
	err := helpers.ValidateWithShouldBind(c, &payload)
	if err != nil {
		return
	}

	userDetail := helpers.GetAuthUser(c)

	var results models.UserFollowings
	filter := bson.D{
		{Key: "user_followings_user", Value: userDetail.UsersId},
		{Key: "user_followings_following", Value: payload.UserFollowingsFollowing},
	}
	checkUserErr := config.DB.Collection("UserFollowings").FindOne(context.TODO(), filter).Decode(&results)

	var FollowingUser models.Users
	followingUserErr := OosaUserRepository{}.ReadUserById(payload.UserFollowingsFollowing, &FollowingUser)

	if followingUserErr == mongo.ErrNoDocuments {
		helpers.ResponseError(c, "Invalid user to follow")
		return
	}

	if checkUserErr != nil {
		countUserCurrent := uf.CountUserFollowing(c, userDetail.UsersId)
		countUserFollowing := uf.CountUserFollowing(c, payload.UserFollowingsFollowing)
		friendListLimit := config.APP_LIMIT.FriendListLimit

		if countUserCurrent+1 > friendListLimit {
			helpers.ResponseError(c, "You cannot add more friends as it has exceeded the allowed limit of "+strconv.Itoa(int(friendListLimit)))
			return
		}
		if countUserFollowing+1 > friendListLimit {
			helpers.ResponseError(c, "You cannot add "+FollowingUser.UsersName+" as friend as his list has exceeded "+strconv.Itoa(int(friendListLimit)))
			return
		}

		insert := models.UserFollowings{
			UserFollowingsUser:      userDetail.UsersId,
			UserFollowingsFollowing: payload.UserFollowingsFollowing,
			UserFollowingsCreatedAt: primitive.NewDateTimeFromTime(time.Now()),
		}

		result, _ := config.DB.Collection("UserFollowings").InsertOne(context.TODO(), insert)
		requestId := result.InsertedID.(primitive.ObjectID)

		insert.UserFollowingsId = requestId
		NotificationMessage := models.NotificationMessage{
			Message: "{0}發送了好友邀請給你!",
			Data:    []map[string]interface{}{helpers.NotificationFormatUser(userDetail), helpers.NotificationFormatUserFollowing(insert)},
		}
		helpers.NotificationsCreate(c, helpers.NOTIFICATION_FRIEND_REQUEST, payload.UserFollowingsFollowing, NotificationMessage, requestId)
		c.JSON(http.StatusOK, gin.H{"message": "User following created successfully", "inserted_id": result.InsertedID})

	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Already followed"})
	}
	uf.CountFollowingFollowers(userDetail.UsersId)
	uf.CountFollowingFollowers(FollowingUser.UsersId)
}

func (uf UserRepository) CountUserFollowing(c *gin.Context, userId primitive.ObjectID) int64 {
	opts := options.Count().SetHint("_id_")
	filter := bson.D{{Key: "user_followings_user", Value: userId}}
	countUserFollowing, _ := config.DB.Collection("UserFollowings").CountDocuments(context.TODO(), filter, opts)
	return countUserFollowing
}

// @Summary UserFollowings
// @Description Get specific userfollowings data
// @ID userfollowings
// @Produce json
// @Tags UserFollowings
// @Success 200 {object} models.UserFollowings
// @Failure 400 {object} structs.Message
// @Router /userfollowings [get]
func (uf UserRepository) UserFollowingRead(c *gin.Context) {
	var userFollowings models.UserFollowings
	err := uf.ReadOne(c, &userFollowings, "")

	if err != nil {
		if err == mongo.ErrNoDocuments {
			helpers.ResponseNoData(c, err.Error())
			return
		}
	}

	c.JSON(200, userFollowings)
}

// @Summary UserFollowings
// @Description Update userfollowings data
// @ID userfollowings
// @Produce json
// @Tags UserFollowings
// @Success 200 {object} Message and models.UserFollowings
// @Failure 400 {object} structs.Message
// @Router /userfollowings [put]
func (uf UserRepository) UserFollowingUpdate(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	var payload models.UserFollowings
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userFollowings models.UserFollowings
	err := config.DB.Collection("UserFollowings").FindOne(context.TODO(), bson.D{{Key: "_id", Value: id}}).Decode(&userFollowings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			helpers.ResponseNoData(c, err.Error())
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userFollowings.UserFollowingsFollowing = payload.UserFollowingsFollowing

	_, errUpd := config.DB.Collection("UserFollowings").ReplaceOne(context.TODO(), bson.D{{Key: "_id", Value: userFollowings.UserFollowingsId}}, userFollowings)
	if errUpd != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uf.CountFollowingFollowers(userFollowings.UserFollowingsUser)
	uf.CountFollowingFollowers(userFollowings.UserFollowingsFollowing)
	c.JSON(http.StatusOK, gin.H{"message": "User following updated successfully"})
}

// @Summary UserFollowings
// @Description Delete userfollowings data
// @ID userfollowings
// @Produce json
// @Tags UserFollowings
// @Success 200 {object} structs.Message
// @Failure 400 {object} structs.Message
// @Router /userfollowings [put]
func (uf UserRepository) UserFollowingDelete(c *gin.Context) {
	var UserFollowings models.UserFollowings
	err := uf.ReadOne(c, &UserFollowings, "")

	if err != nil {
		if err == mongo.ErrNoDocuments {
			helpers.ResponseNoData(c, err.Error())
			return
		}
	}

	_, errDelete := config.DB.Collection("UserFollowings").DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: UserFollowings.UserFollowingsId}})
	if errDelete != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uf.CountFollowingFollowers(UserFollowings.UserFollowingsUser)
	uf.CountFollowingFollowers(UserFollowings.UserFollowingsFollowing)
	c.JSON(http.StatusOK, gin.H{"message": "User following deleted successfully"})
}

func (uf UserRepository) RetrieveUsers(c *gin.Context) {
	userName := c.Query("name")
	//userDetail := helpers.GetAuthUser(c)

	var results []models.UsersAgg

	filter := bson.D{}

	if userName != "" {
		filter = append(filter, bson.E{Key: "users_name", Value: bson.M{"$regex": userName, "$options": "i"}})
	}

	cursor, _ := config.DB.Collection("Users").Find(context.TODO(), filter)
	cursor.All(context.TODO(), &results)
	c.JSON(http.StatusOK, results)
}

func (uf UserRepository) ReadOne(c *gin.Context, UserFollowings *models.UserFollowings, userFollowingId string) error {
	userDetail := helpers.GetAuthUser(c)
	idVal := c.Param("id")

	if userFollowingId != "" {
		idVal = userFollowingId
	}

	id, _ := primitive.ObjectIDFromHex(idVal)

	filter := bson.D{
		{Key: "_id", Value: id},
		{Key: "user_followings_user", Value: userDetail.UsersId},
	}

	err := config.DB.Collection("UserFollowings").FindOne(context.TODO(), filter).Decode(&UserFollowings)
	return err
}

func (uf UserRepository) CountFollowingFollowers(userId primitive.ObjectID) {
	following, _ := config.DB.Collection("UserFollowings").CountDocuments(context.TODO(), bson.D{{Key: "user_followings_user", Value: userId}})
	follower, _ := config.DB.Collection("UserFollowings").CountDocuments(context.TODO(), bson.D{{Key: "user_followings_following", Value: userId}})

	update := bson.D{{Key: "$set", Value: bson.M{
		"users_following_count": int(following),
		"users_follower_count":  int(follower),
	}}}
	config.DB.Collection("Users").UpdateOne(context.TODO(), bson.D{{Key: "_id", Value: userId}}, update)
}
