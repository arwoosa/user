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
		c.JSON(http.StatusOK, "Already followed")
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

// @Summary RetrieveUserSettings
// @Description Retrieve user settings
// @ID RetrieveUserSetting
// @Produce json
// @Tags UserFollowings
// @Success 200 {object} User Settings
// @Failure 400 {object} structs.Message
// @Router /userfollowings [get]
func (uf UserRepository) RetrieveUserSettings(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{"_id": id},
		},
		{
			"$project": bson.M{
				"_id":              0,
				"users_source":     0,
				"users_avatar":     0,
				"users_source_id":  0,
				"users_name":       0,
				"users_email":      0,
				"users_password":   0,
				"users_object":     0,
				"users_created_at": 0,
			},
		},
	}

	cursor, err := config.DB.Collection("Users").Aggregate(context.TODO(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var results []bson.M
	if err := cursor.All(context.TODO(), &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(results) == 0 {
		helpers.ResponseNoData(c, "No user found")
		return
	}

	c.JSON(http.StatusOK, results)
}

// @Summary RetrieveUserSettings
// @Description Retrieve user settings
// @ID RetrieveUserSetting
// @Produce json
// @Tags UserFollowings
// @Success 200 {object} User Settings
// @Failure 400 {object} structs.Message
// @Router /userfollowings [get]
func (uf UserRepository) UpdateUserSettings(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var updateData models.Users
	if err := c.BindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.M{"_id": id}
	updateFields := bson.M{
		"users_setting_vis_events":              updateData.UsersSettingVisEvents,
		"users_setting_vis_achievement_journal": updateData.UsersSettingVisAchievementJournal,
		"users_setting_vis_collab_log":          updateData.UsersSettingVisCollabLog,
		"users_setting_vis_follow":              updateData.UsersSettingVisFollow,
	}

	var updatedUser models.Users
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = config.DB.Collection("Users").FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": updateFields}, options).Decode(&updatedUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"_id":                                   updatedUser.UsersId,
		"users_setting_vis_events":              updatedUser.UsersSettingVisEvents,
		"users_setting_vis_achievement_journal": updatedUser.UsersSettingVisAchievementJournal,
		"users_setting_vis_collab_log":          updatedUser.UsersSettingVisCollabLog,
		"users_setting_vis_follow":              updatedUser.UsersSettingVisFollow,
	})
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
