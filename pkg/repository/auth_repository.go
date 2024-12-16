package repository

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"
	"strconv"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/api/idtoken"
)

type AuthRepository struct{}

type AuthGoogleRequest struct {
	Credential string `json:"credential"`
}
type AuthUpdateRequest struct {
	UsersName            string `json:"users_name"`
	UsersUsername        string `json:"users_username"`
	UsersEmail           string `json:"users_email"`
	UsersSettingLanguage string `json:"users_setting_language"`
}

type AuthUpdateTakeMeRequest struct {
	UsersTakeMeStatus int `json:"users_take_me_status"`
}
type AuthUpdatePasswordRequest struct {
	Password    string `json:"password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type AuthUpdateAvatarRequest struct {
	UsersAvatar string `json:"users_avatar" validate:"required"`
}

/*
	User Source
	1	Google
	2	Line
	3	Email
	4	Facebook
*/

// AuthGoogle handles Google authentication.
// @Summary Authenticate (Google)
// @Description Authenticate (Google)
// @ID authenticate-google
// @Produce json
// @Tags Authentication
// @Param Request body AuthGoogleRequest true "Request Parmeter"
// @Success 200 {object} AuthOOSA
// @Failure 400 {object} structs.Message
// @Router /auth/google [post]
func (t AuthRepository) AuthGoogle(c *gin.Context) {
	var payload AuthGoogleRequest
	helpers.Validate(c, &payload)
	googlePayload, errGoogleAuth := idtoken.Validate(c, payload.Credential, config.APP.OauthGoogleClientId)
	if errGoogleAuth != nil {
		helpers.ResponseNoData(c, errGoogleAuth.Error())
		return
	}

	var User models.Users
	errUser := config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "users_source", Value: 1}, {Key: "users_source_id", Value: googlePayload.Subject}}).Decode(&User)

	if errUser != nil {
		if errUser == mongo.ErrNoDocuments {
			autoAdd := 0
			takeMeStatus := false
			insert := models.Users{
				UsersSource:                           1,
				UsersSourceId:                         googlePayload.Subject,
				UsersName:                             googlePayload.Claims["name"].(string),
				UsersObject:                           googlePayload.Subject,
				UsersAvatar:                           googlePayload.Claims["picture"].(string),
				UsersSettingLanguage:                  "",
				UsersSettingIsVisibleFriends:          1,
				UsersSettingIsVisibleStatistics:       1,
				UsersSettingVisibilityActivitySummary: 1,
				UsersSettingFriendAutoAdd:             &autoAdd,
				UsersTakeMeStatus:                     &takeMeStatus,
				UsersIsSubscribed:                     false,
				UsersCreatedAt:                        primitive.NewDateTimeFromTime(time.Now()),
			}
			result, _ := config.DB.Collection("Users").InsertOne(context.TODO(), insert)
			config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "_id", Value: result.InsertedID}}).Decode(&User)
		} else {
			helpers.ResponseNoData(c, errUser.Error())
			return
		}
	}
	helpers.AuthenticateUser(c, User)
}

// Auth handles authentication.
// @Summary Auth Detail
// @Description Get the current logged in detail
// @ID authenticate-read
// @Produce json
// @Tags Authentication
// @Success 200 {object} models.Users
// @Failure 400 {object} structs.Message
// @Security ApiKeyAuth
// @Router /auth [get]
func (t AuthRepository) Auth(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)
	userDetail.UsersBreathingPoints = t.Breathing(c, userDetail.UsersId)
	c.JSON(http.StatusOK, userDetail)
}

func (t AuthRepository) AuthUpdateTakeMe(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)
	var payload AuthUpdateTakeMeRequest

	validateError := helpers.Validate(c, &payload)
	if validateError != nil {
		return
	}

	UpdateUser := models.Users{}

	status := false
	if payload.UsersTakeMeStatus == 1 {
		status = true
	}
	UpdateUser.UsersTakeMeStatus = &status

	filters := bson.D{{Key: "_id", Value: userDetail.UsersId}}

	var User models.Users
	upd := bson.D{{Key: "$set", Value: UpdateUser}}
	config.DB.Collection("Users").UpdateOne(context.TODO(), filters, upd)
	config.DB.Collection("Users").FindOne(context.TODO(), filters).Decode(&User)

	c.JSON(http.StatusOK, User)
}

func (t AuthRepository) AuthUpdate(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)
	var payload AuthUpdateRequest

	validateError := helpers.Validate(c, &payload)
	if validateError != nil {
		return
	}

	UpdateUser := models.Users{}

	if payload.UsersUsername != "" {
		match := helpers.RegexCompare(helpers.REGEX_USERNAME, payload.UsersUsername)
		if !match {
			helpers.ResponseBadRequestError(c, "Unable to change username as it does not fulfill criteria")
			return
		}

		// Find if username already used
		var User models.Users
		filter := bson.D{{Key: "users_username", Value: payload.UsersUsername}}
		config.DB.Collection("Users").FindOne(context.TODO(), filter).Decode(&User)

		currentTime := time.Now()
		lastUpdate := userDetail.UsersUsernameLastUpdate.Time()

		currentTimeCompare := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 23, 59, 59, 0, currentTime.Location())
		lastUpdateCompare := time.Date(lastUpdate.Year(), lastUpdate.Month(), lastUpdate.Day(), 23, 59, 59, 0, lastUpdate.Location())

		days := int(math.Ceil(currentTimeCompare.Sub(lastUpdateCompare).Hours() / 24))

		if days < 60 {
			helpers.ResponseBadRequestError(c, "Unable to change username as it was changed "+strconv.Itoa(days)+" day ago. We only allow a change every 60 days")
			return
		}

		if helpers.MongoZeroID(User.UsersId) {
			UpdateUser.UsersUsername = payload.UsersUsername
			UpdateUser.UsersUsernameLastUpdate = primitive.NewDateTimeFromTime(time.Now())
		} else {
			if userDetail.UsersId != User.UsersId {
				helpers.ResponseError(c, "Username already used")
				return
			}
			// helpers.ResponseSuccessMessage(c, "Username did not change")
			// return
		}
	}

	if payload.UsersEmail != "" {
		// Find if email already used
		var User models.Users
		filter := bson.D{{Key: "users_email", Value: payload.UsersEmail}}
		config.DB.Collection("Users").FindOne(context.TODO(), filter).Decode(&User)

		if helpers.MongoZeroID(User.UsersId) {
			UpdateUser.UsersEmail = payload.UsersEmail
		} else {
			if userDetail.UsersId != User.UsersId {
				helpers.ResponseError(c, "Email already used")
				return
			}
			// helpers.ResponseSuccessMessage(c, "Email did not change")
			// return
		}
	}

	if payload.UsersName != "" {
		match := helpers.RegexCompare(helpers.REGEX_NAME, payload.UsersName)
		if !match {
			helpers.ResponseBadRequestError(c, "Unable to change name as it does not fulfill criteria")
			return
		}

		currentTime := time.Now()
		lastUpdate := userDetail.UsersNameLastUpdate.Time()

		currentTimeCompare := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 23, 59, 59, 0, currentTime.Location())
		lastUpdateCompare := time.Date(lastUpdate.Year(), lastUpdate.Month(), lastUpdate.Day(), 23, 59, 59, 0, lastUpdate.Location())

		days := int(math.Ceil(currentTimeCompare.Sub(lastUpdateCompare).Hours() / 24))

		if days < 60 {
			helpers.ResponseBadRequestError(c, "Unable to change username as it was changed "+strconv.Itoa(days)+" day ago. We only allow a change every 60 days")
			return
		}

		UpdateUser.UsersNameLastUpdate = primitive.NewDateTimeFromTime(time.Now())
		UpdateUser.UsersName = payload.UsersName
	}

	if payload.UsersSettingLanguage != "" {
		UpdateUser.UsersSettingLanguage = payload.UsersSettingLanguage
	}

	filters := bson.D{{Key: "_id", Value: userDetail.UsersId}}

	var User models.Users
	upd := bson.D{{Key: "$set", Value: UpdateUser}}
	config.DB.Collection("Users").UpdateOne(context.TODO(), filters, upd)
	config.DB.Collection("Users").FindOne(context.TODO(), filters).Decode(&User)

	c.JSON(http.StatusOK, User)
}

func (t AuthRepository) AuthUpdatePassword(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)
	var User models.Users
	var payload AuthUpdatePasswordRequest

	validateError := helpers.Validate(c, &payload)
	if validateError != nil {
		return
	}

	config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "_id", Value: userDetail.UsersId}}).Decode(&User)

	match := helpers.RegexCompare(helpers.REGEX_PASSWORD, payload.NewPassword)
	if !match {
		helpers.ResponseBadRequestError(c, "Password does not fulfill criteria")
		return
	}

	checkCredential := helpers.CheckPassword(User.UsersPassword, payload.Password)
	if !checkCredential {
		helpers.ResponseError(c, "Unable to change password as old password does not match")
		return
	}

	hashedPassword, err := helpers.HashPassword(payload.NewPassword)
	if err != nil {
		helpers.ResponseError(c, "Failed to hash password")
		return
	}

	filters := bson.D{{Key: "_id", Value: userDetail.UsersId}}
	UpdateUser := models.Users{
		UsersPassword: hashedPassword,
	}
	upd := bson.D{{Key: "$set", Value: UpdateUser}}
	config.DB.Collection("Users").UpdateOne(context.TODO(), filters, upd)

	// c.JSON(200, userDetail)
	c.JSON(http.StatusOK, gin.H{"message": "Password updated"})
}

func (t AuthRepository) AuthUpdateAvatar(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)
	var User models.Users
	var payload AuthUpdateAvatarRequest

	validateError := helpers.Validate(c, &payload)
	if validateError != nil {
		return
	}

	config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "_id", Value: userDetail.UsersId}}).Decode(&User)
	filters := bson.D{{Key: "_id", Value: userDetail.UsersId}}
	userDetail.UsersAvatar = payload.UsersAvatar
	UpdateUser := models.Users{
		UsersAvatar: payload.UsersAvatar,
	}
	upd := bson.D{{Key: "$set", Value: UpdateUser}}
	config.DB.Collection("Users").UpdateOne(context.TODO(), filters, upd)

	c.JSON(200, userDetail)
}

func (t AuthRepository) AuthUpdateProfilePicture(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)

	file, fileErr := c.FormFile("users_avatar")
	if fileErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "No file is received",
		})
		return
	}

	uploadedFile, err := file.Open()

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Unable to open file",
		})
		return
	}

	b, _ := io.ReadAll(uploadedFile)
	mimeType := mimetype.Detect(b)

	switch mimeType.String() {
	case "image/jpeg":
	case "image/png":
	default:
		c.JSON(http.StatusBadRequest, "Mime: "+mimeType.String()+" not supported")
		return
	}

	cloudflare := CloudflareRepository{}
	cloudflareResponse, postErr := cloudflare.Post(c, file)
	if postErr != nil {
		helpers.ResponseBadRequestError(c, postErr.Error())
		return
	}
	fileName := cloudflare.ImageDelivery(cloudflareResponse.Result.Id, "public")

	filters := bson.D{{Key: "_id", Value: userDetail.UsersId}}
	upd := bson.D{{Key: "$set", Value: models.Users{
		UsersAvatar: fileName,
	}}}
	userDetail.UsersAvatar = fileName
	config.DB.Collection("Users").UpdateOne(context.TODO(), filters, upd)

	c.JSON(http.StatusOK, userDetail)
}

// AuthLine handles Line authentication.
// @Summary Authenticate (Line)
// @Description Authenticate (Line)
// @ID authenticate-line
// @Produce json
// @Tags Authentication
// @Param Request body AuthLineRequest true "Request Parameter"
// @Success 200 {object} AuthOOSA
// @Failure 400 {object} structs.Message
// @Router /auth/line [post]
// @Code Roy
func (t AuthRepository) AuthLine(c *gin.Context) {
	var params helpers.AuthLineRequest
	if err := c.ShouldBind(&params); err != nil {
		helpers.ResponseBadRequestError(c, fmt.Sprintf("failed to bind request: %s", err.Error()))
		return
	}

	accessToken, err := helpers.GetLineAccessToken(params)
	if err != nil {
		helpers.ResponseError(c, fmt.Sprintf("GetLineAccessToken Error: %s", err.Error()))
		return
	}

	userInfo, err := helpers.GetUserInfo(accessToken.AccessToken)
	if err != nil {
		helpers.ResponseError(c, fmt.Sprintf("getUserInfo Error: %s", err.Error()))
		return
	}

	var User models.Users
	errUser := config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "users_source", Value: 2}, {Key: "users_source_id", Value: userInfo.UserID}}).Decode(&User)

	if errUser != nil {
		if errUser == mongo.ErrNoDocuments {
			autoAdd := 0
			takeMeStatus := false
			insert := models.Users{
				UsersSource:                           2,
				UsersSourceId:                         userInfo.UserID,
				UsersName:                             userInfo.Name,
				UsersObject:                           userInfo.UserID,
				UsersAvatar:                           userInfo.Picture,
				UsersSettingLanguage:                  "",
				UsersSettingIsVisibleFriends:          1,
				UsersSettingIsVisibleStatistics:       1,
				UsersSettingVisibilityActivitySummary: 1,
				UsersSettingFriendAutoAdd:             &autoAdd,
				UsersTakeMeStatus:                     &takeMeStatus,
				UsersIsSubscribed:                     false,
				UsersCreatedAt:                        primitive.NewDateTimeFromTime(time.Now()),
			}
			result, _ := config.DB.Collection("Users").InsertOne(context.TODO(), insert)
			config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "_id", Value: result.InsertedID}}).Decode(&User)
		} else {
			helpers.ResponseNoData(c, errUser.Error())
			return
		}
	}
	helpers.AuthenticateUser(c, User)
}

// AuthFacebook handles Facebook authentication.
// @Summary Authenticate (Facebook)
// @Description Authenticate (Facebook)
// @ID authenticate-Facebook
// @Produce json
// @Tags Authentication
// @Param Request body AuthFacebookRequest true "Request Parameter"
// @Success 200 {object} AuthOOSA
// @Failure 400 {object} structs.Message
// @Router /auth/Facebook [post]
// @Code Roy
func (t AuthRepository) AuthFacebook(c *gin.Context) {
	var payload helpers.AuthFacebookRequest

	if err := c.ShouldBind(&payload); err != nil {
		helpers.ResponseBadRequestError(c, "Empty request body")
		return
	}

	var User models.Users
	errUser := config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "users_source", Value: 4}, {Key: "users_source_id", Value: payload.Id}}).Decode(&User)

	if errUser != nil {
		if errUser == mongo.ErrNoDocuments {
			autoAdd := 0
			takeMeStatus := false
			insert := models.Users{
				UsersSource:                           4,
				UsersSourceId:                         payload.Id,
				UsersName:                             payload.Name,
				UsersObject:                           payload.Id,
				UsersAvatar:                           "",
				UsersSettingLanguage:                  "",
				UsersSettingIsVisibleFriends:          1,
				UsersSettingIsVisibleStatistics:       1,
				UsersSettingVisibilityActivitySummary: 1,
				UsersSettingFriendAutoAdd:             &autoAdd,
				UsersTakeMeStatus:                     &takeMeStatus,
				UsersIsSubscribed:                     false,
				UsersCreatedAt:                        primitive.NewDateTimeFromTime(time.Now()),
			}
			result, _ := config.DB.Collection("Users").InsertOne(context.TODO(), insert)
			config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "_id", Value: result.InsertedID}}).Decode(&User)
		} else {
			helpers.ResponseNoData(c, errUser.Error())
			return
		}
	}
	helpers.AuthenticateUser(c, User)
}

// RegisterEmail handles email registration.
// @Summary Register (Email)
// @Description Register (Email)
// @ID register-Email
// @Produce json
// @Tags Authentication
// @Param Request body AuthEmailRequest true "Request Parameter"
// @Success 200 {object} AuthOOSA
// @Failure 400 {object} structs.Message
// @Router /auth/email [post]
// @Code Roy
func (t AuthRepository) RegisterEmail(c *gin.Context) {
	var payload helpers.AuthEmailRequest

	validateError := helpers.Validate(c, &payload)
	if validateError != nil {
		return
	}

	var User models.Users
	errUser := config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "users_source", Value: 3}, {Key: "users_email", Value: payload.Email}}).Decode(&User)

	match := helpers.RegexCompare(helpers.REGEX_PASSWORD, payload.Password)
	if !match {
		helpers.ResponseBadRequestError(c, "Password does not fulfill criteria")
		return
	}

	matchName := helpers.RegexCompare(helpers.REGEX_NAME, payload.Name)
	if !matchName {
		helpers.ResponseBadRequestError(c, "Unable to change name as it does not fulfill criteria")
		return
	}

	hashedPassword, err := helpers.HashPassword(payload.Password)
	if err != nil {
		helpers.ResponseError(c, "Failed to hash password")
		return
	}

	isBusiness := false

	if payload.IsBusiness {
		isBusiness = true
	}

	if errUser != nil {
		if errUser == mongo.ErrNoDocuments {
			newUUID := uuid.New()
			uuid := newUUID.String()
			autoAdd := 0
			takeMeStatus := false
			insert := models.Users{
				UsersSource:                           3,
				UsersSourceId:                         uuid,
				UsersName:                             payload.Name,
				UsersEmail:                            payload.Email,
				UsersPassword:                         hashedPassword,
				UsersObject:                           uuid,
				UsersAvatar:                           "",
				UsersSettingLanguage:                  "",
				UsersSettingIsVisibleFriends:          1,
				UsersSettingIsVisibleStatistics:       1,
				UsersSettingVisibilityActivitySummary: 1,
				UsersSettingFriendAutoAdd:             &autoAdd,
				UsersTakeMeStatus:                     &takeMeStatus,
				UsersIsSubscribed:                     false,
				UsersIsBusiness:                       isBusiness,
				UsersCreatedAt:                        primitive.NewDateTimeFromTime(time.Now()),
			}
			result, _ := config.DB.Collection("Users").InsertOne(context.TODO(), insert)
			config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "_id", Value: result.InsertedID}}).Decode(&User)
		} else {
			helpers.ResponseNoData(c, errUser.Error())
			return
		}
		helpers.AuthenticateUser(c, User)
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "email already registered", "data": nil})
	}
}

// AuthEmail handles email authentication.
// @Auth Email (Email)
// @Description Auth (Email)
// @ID register-Email
// @Produce json
// @Tags Authentication
// @Param Request body AuthEmail true AuthEmailRequest
// @Success 200 {object} RegisterEmail
// @Failure 400 {object} structs.Message
// @Router /auth/email [post]
// @Code Roy
func (t AuthRepository) AuthEmail(c *gin.Context) {
	var payload helpers.AuthEmailRequest

	if err := c.ShouldBind(&payload); err != nil {
		helpers.ResponseBadRequestError(c, "Empty request body")
		return
	}

	var User models.Users
	errUser := config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "users_source", Value: 3}, {Key: "users_email", Value: payload.Email}}).Decode(&User)

	if errUser != nil {
		if errUser == mongo.ErrNoDocuments {
			c.JSON(http.StatusOK, gin.H{"message": "no credentials found", "data": nil})
			return
		} else {
			helpers.ResponseError(c, errUser.Error())
			return
		}
	}

	checkCredential := helpers.CheckPassword(User.UsersPassword, payload.Password)
	if !checkCredential {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials", "data": nil})
		return
	}

	helpers.AuthenticateUser(c, User)
}

// @Summary RetrieveUserSettings
// @Description Retrieve user settings
// @ID RetrieveUserSetting
// @Produce json
// @Tags UserFollowings
// @Success 200 {object} User Settings
// @Failure 400 {object} structs.Message
// @Router /userfollowings [get]
func (t AuthRepository) RetrieveUserSettings(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)

	pipeline := []bson.M{
		{
			"$match": bson.M{"_id": userDetail.UsersId},
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

	c.JSON(http.StatusOK, results[0])
}

var PrivacySettings = map[int]string{
	1: "Present to whole user",
	2: "Only followed user",
	3: "Only user itself",
}

// @Summary RetrieveUserSettings
// @Description Retrieve user settings
// @ID RetrieveUserSetting
// @Produce json
// @Tags UserFollowings
// @Success 200 {object} User Settings
// @Failure 400 {object} structs.Message
// @Router /userfollowings [get]
func (t AuthRepository) UpdateUserSettings(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)

	var updateData models.Users
	if err := c.BindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.M{"_id": userDetail.UsersId}
	updateFields := bson.M{
		"users_setting_is_visible_friends":          updateData.UsersSettingIsVisibleFriends,
		"users_setting_is_visible_statistics":       updateData.UsersSettingIsVisibleStatistics,
		"users_setting_visibility_activity_summary": updateData.UsersSettingVisibilityActivitySummary,
		"users_setting_friend_auto_add":             updateData.UsersSettingFriendAutoAdd,
	}
	fmt.Println("TEST")
	var updatedUser models.Users
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := config.DB.Collection("Users").FindOneAndUpdate(context.TODO(), filter, bson.M{"$set": updateFields}, options).Decode(&updatedUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

func (t AuthRepository) RetrieveBadges(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)
	var results []models.Badges
	err := t.RetrieveBadgesByUserId(c, userDetail.UsersId, &results)

	if err != nil {
		return
	}

	c.JSON(http.StatusOK, results)
}

func (t AuthRepository) RetrieveBadgesByUserId(c *gin.Context, userId primitive.ObjectID, Badges *[]models.Badges) error {
	single := c.Query("single")
	agg := mongo.Pipeline{
		bson.D{{
			Key: "$match", Value: bson.M{"user_badges_user": userId},
		}},
		bson.D{{
			Key: "$lookup", Value: bson.M{
				"from":         "Badges",
				"localField":   "user_badges_badge",
				"foreignField": "_id",
				"as":           "UserBadgesDetail",
			},
		}},
		bson.D{{
			Key: "$unwind", Value: bson.M{"path": "$UserBadgesDetail"},
		}},
		bson.D{{
			Key: "$replaceRoot", Value: bson.M{"newRoot": "$UserBadgesDetail"},
		}},
	}

	if single != "" && single == "true" {
		agg = append(agg,
			bson.D{{
				Key: "$match", Value: bson.M{"UserBadgesDetail.badges_is_once": true},
			}})
	} else if single != "" && single == "false" {
		agg = append(agg,
			bson.D{{
				Key: "$match", Value: bson.M{"UserBadgesDetail.badges_is_once": false},
			}})
	}
	cursor, err := config.DB.Collection("UserBadges").Aggregate(context.TODO(), agg)
	cursor.All(context.TODO(), Badges)

	for k, v := range *Badges {
		(*Badges)[k].BadgesUrl = config.APP.BaseUrl + "badges/" + v.BadgesCode + ".png"
	}

	return err
}

func (t AuthRepository) RetrieveNotifications(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)
	var results []models.Notifications

	filter := bson.D{
		{Key: "notifications_user", Value: userDetail.UsersId},
	}
	cursor, err := config.DB.Collection("Notifications").Find(context.TODO(), filter)

	if err != nil {
		helpers.ResponseBadRequestError(c, err.Error())
		return
	}
	cursor.All(context.TODO(), &results)

	c.JSON(200, results)
}

func (t AuthRepository) BindFacebook(c *gin.Context) {
	userDetail := helpers.GetAuthUser(c)
	var payload helpers.AuthBindingFacebookRequest

	if err := c.ShouldBind(&payload); err != nil {
		helpers.ResponseBadRequestError(c, "Empty request body")
		return
	}

	oauth := FacebookRepository{}.Retrieve(c, payload.AccessToken)

	if oauth.Id != "" {
		userDetail.UsersBindingFacebook = payload.Id
		filters := bson.D{{Key: "_id", Value: userDetail.UsersId}}
		upd := bson.D{{Key: "$set", Value: userDetail}}
		config.DB.Collection("Users").UpdateOne(context.TODO(), filters, upd)

		helpers.BadgeAllocate(c, "N5", helpers.BADGE_SOCIAL, primitive.NilObjectID, primitive.NilObjectID)
		c.JSON(200, userDetail)
	} else {
		c.JSON(400, gin.H{"message": "Invalid facebook token"})
	}

}

func GetEventParticipantStatus(status string) int64 {
	ParticipantStatus := map[string]int64{
		"PENDING":  0,
		"ACCEPTED": 1,
		"REJECTED": 2,
	}
	return ParticipantStatus[status]
}

func (t AuthRepository) Breathing(c *gin.Context, userId primitive.ObjectID) int {
	var results []models.Events

	agg := mongo.Pipeline{
		bson.D{{
			Key: "$match", Value: bson.M{
				"event_participants_user":   userId,
				"event_participants_status": GetEventParticipantStatus("ACCEPTED"),
			},
		}},
		bson.D{{
			Key: "$lookup", Value: bson.M{
				"from":         "Events",
				"localField":   "event_participants_event",
				"foreignField": "_id",
				"as":           "event_participants_event_detail",
			},
		}},
		bson.D{{
			Key: "$unwind", Value: "$event_participants_event_detail",
		}},
		bson.D{{
			Key: "$replaceRoot", Value: bson.M{
				"newRoot": "$event_participants_event_detail",
			},
		}},
		bson.D{{
			Key: "$sort", Value: bson.M{
				"events_date": -1,
			},
		}},
		bson.D{{
			Key: "$limit", Value: 1,
		}},
	}
	cursor, err := config.DB.Collection("EventParticipants").Aggregate(context.TODO(), agg)
	cursor.All(context.TODO(), &results)

	if err != nil {
		return 0
	}

	if len(results) == 0 {
		return 0
	}

	lastEventDate := helpers.MongoTimestampToTime(results[0].EventsDate)
	currentTime := time.Now()
	diff := currentTime.Sub(lastEventDate)
	numberOfDays := diff.Hours() / 24

	breathingPoints := 100 - int(math.Ceil(numberOfDays))

	var Exp []models.Exp
	expAgg := mongo.Pipeline{
		bson.D{{
			Key: "$match", Value: bson.M{
				"exp_user": userId,
			},
		}},
		bson.D{{
			Key: "$group", Value: bson.M{
				"_id":        "$exp_user",
				"exp_points": bson.M{"$sum": "$exp_points"},
			},
		}},
	}
	cursorExp, _ := config.DB.Collection("Exp").Aggregate(context.TODO(), expAgg)
	cursorExp.All(context.TODO(), &Exp)

	if len(Exp) > 0 {
		breathingPoints += Exp[0].ExpPoints
	}

	if breathingPoints > 100 {
		breathingPoints = 100
	} else if breathingPoints < 0 {
		breathingPoints = 0
	}

	return breathingPoints
}
