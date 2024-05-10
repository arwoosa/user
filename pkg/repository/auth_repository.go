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
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/idtoken"
)

type AuthRepository struct{}

type AuthGoogleRequest struct {
	Credential string `json:"credential"`
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
			insert := models.Users{
				UsersSource:                       1,
				UsersSourceId:                     googlePayload.Subject,
				UsersName:                         googlePayload.Claims["name"].(string),
				UsersObject:                       googlePayload.Subject,
				UsersAvatar:                       googlePayload.Claims["picture"].(string),
				UsersSettingLanguage:              "",
				UsersSettingVisEvents:             1,
				UsersSettingVisAchievementJournal: 1,
				UsersSettingVisCollabLog:          1,
				UsersSettingVisFollow:             1,
				UsersIsSubscribed:                 false,
				UsersCreatedAt:                    primitive.NewDateTimeFromTime(time.Now()),
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
	user, _ := c.Get("user")
	c.JSON(http.StatusOK, user)
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
			insert := models.Users{
				UsersSource:                       2,
				UsersSourceId:                     userInfo.UserID,
				UsersName:                         userInfo.Name,
				UsersObject:                       userInfo.UserID,
				UsersAvatar:                       userInfo.Picture,
				UsersSettingLanguage:              "",
				UsersSettingVisEvents:             1,
				UsersSettingVisAchievementJournal: 1,
				UsersSettingVisCollabLog:          1,
				UsersSettingVisFollow:             1,
				UsersIsSubscribed:                 false,
				UsersCreatedAt:                    primitive.NewDateTimeFromTime(time.Now()),
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
			insert := models.Users{
				UsersSource:                       4,
				UsersSourceId:                     payload.Id,
				UsersName:                         payload.Name,
				UsersObject:                       payload.Id,
				UsersAvatar:                       "",
				UsersSettingLanguage:              "",
				UsersSettingVisEvents:             1,
				UsersSettingVisAchievementJournal: 1,
				UsersSettingVisCollabLog:          1,
				UsersSettingVisFollow:             1,
				UsersIsSubscribed:                 false,
				UsersCreatedAt:                    primitive.NewDateTimeFromTime(time.Now()),
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

	hashedPassword, err := helpers.HashPassword(payload.Password)
	if err != nil {
		helpers.ResponseError(c, "Failed to hash password")
		return
	}

	if errUser != nil {
		if errUser == mongo.ErrNoDocuments {
			newUUID := uuid.New()
			uuid := newUUID.String()
			insert := models.Users{
				UsersSource:                       3,
				UsersSourceId:                     uuid,
				UsersName:                         payload.Name,
				UsersEmail:                        payload.Email,
				UsersPassword:                     hashedPassword,
				UsersObject:                       uuid,
				UsersAvatar:                       "",
				UsersSettingLanguage:              "",
				UsersSettingVisEvents:             1,
				UsersSettingVisAchievementJournal: 1,
				UsersSettingVisCollabLog:          1,
				UsersSettingVisFollow:             1,
				UsersIsSubscribed:                 false,
				UsersCreatedAt:                    primitive.NewDateTimeFromTime(time.Now()),
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
