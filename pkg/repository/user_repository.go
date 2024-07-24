package repository

import (
	"context"
	"fmt"
	"net/http"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	if checkUserErr != nil {
		insert := models.UserFollowings{
			UserFollowingsUser:      userDetail.UsersId,
			UserFollowingsFollowing: payload.UserFollowingsFollowing,
			UserFollowingsCreatedAt: primitive.NewDateTimeFromTime(time.Now()),
		}

		result, _ := config.DB.Collection("UserFollowings").InsertOne(context.TODO(), insert)
		c.JSON(http.StatusOK, gin.H{"message": "User following created successfully", "inserted_id": result.InsertedID})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Already followed"})
	}

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

	result, err := config.DB.Collection("UserFollowings").ReplaceOne(context.TODO(), bson.D{{Key: "_id", Value: userFollowings.UserFollowingsId}}, userFollowings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(result)
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

	c.JSON(http.StatusOK, gin.H{"message": "User following deleted successfully"})
}

// @Summary UserFollowings
// @Description Retrieve all user friends
// @ID userfollowings
// @Produce json
// @Tags UserFollowings
// @Success 200 {object} []models.UserFollowings
// @Failure 400 {object} structs.Message
// @Router /userfollowings [get]
func (uf UserRepository) RetrieveUserFriends(c *gin.Context) {
	userName := c.Query("name")
	userDetail := helpers.GetAuthUser(c)

	// Define pipeline for aggregation
	pipeline := []bson.M{
		{"$match": bson.M{"user_followings_user": userDetail.UsersId}},
		{"$lookup": bson.M{
			"from":         "Users",
			"localField":   "user_followings_following",
			"foreignField": "_id",
			"as":           "user",
		}},
		{"$unwind": "$user"},
		{"$project": bson.M{
			"users_id":     "$user._id",
			"users_name":   "$user.users_name",
			"users_avatar": "$user.users_avatar",
			"_id":          0,
		}},
	}

	if userName != "" {
		match := bson.M{"$match": bson.M{
			"users_name": bson.M{"$regex": userName, "$options": "i"},
		}}
		pipeline = append(pipeline, match)
	}

	cursor, err := config.DB.Collection("UserFollowings").Aggregate(context.TODO(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user friends"})
		return
	}
	defer cursor.Close(context.TODO())

	var results []bson.M
	if err := cursor.All(context.TODO(), &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user friends data"})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No friends found for the user"})
		return
	}

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
