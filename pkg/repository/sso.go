package repository

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewSSoRepository(url *url.URL) ssoRepository {
	return ssoRepository{
		ssoRegisterUrl: url,
	}
}

type ssoRepository struct {
	ssoRegisterUrl *url.URL
}

var defaultFriendAutoAdd = 1
var defaultFriendTakeMestatus = false

// TODO: add user db
func saveUserInfo(user *UserBindByHeader) (*models.Users, error) {
	var User models.Users

	insert := models.Users{
		UsersSource:                           3,
		UsersSourceId:                         user.Id,
		UsersName:                             user.Name,
		UsersEmail:                            user.Email,
		UsersObject:                           user.User,
		UsersAvatar:                           "",
		UsersSettingLanguage:                  user.Language,
		UsersSettingIsVisibleFriends:          1,
		UsersSettingIsVisibleStatistics:       1,
		UsersSettingVisibilityActivitySummary: 1,
		UsersSettingFriendAutoAdd:             &defaultFriendAutoAdd,
		UsersIsSubscribed:                     false,
		UsersIsBusiness:                       false,
		UsersTakeMeStatus:                     &defaultFriendTakeMestatus,
		UsersCreatedAt:                        primitive.NewDateTimeFromTime(time.Now()),
	}
	result, _ := config.DB.Collection("Users").InsertOne(context.TODO(), insert)
	config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "_id", Value: result.InsertedID}}).Decode(&User)
	return &User, nil
}

func (t ssoRepository) Register(c *gin.Context) {
	if c.Query("return_to") == "" {
		helpers.ResponseBadRequestError(c, "missing return_to")
		return
	}
	registerUrl := *c.Request.URL
	registerUrl.Scheme, registerUrl.Host = getSchemeAndHost(c)
	state := uuid.NewString()
	mysession := sessions.Default(c)
	mysession.Options(sessions.Options{Secure: true, HttpOnly: true, MaxAge: 300, SameSite: http.SameSiteLaxMode, Path: "/"})
	mysession.Set("state", state)
	registerUrl.Path = fmt.Sprintf("/api%s/finish", registerUrl.Path)
	registerUrl.RawQuery = fmt.Sprintf("state=%s", state)

	t.ssoRegisterUrl.RawQuery = fmt.Sprintf("return_to=%s", url.QueryEscape(registerUrl.String()))
	mysession.Set("return_to", c.Query("return_to"))

	mysession.Save()
	c.Redirect(http.StatusSeeOther, t.ssoRegisterUrl.String())
}

func getSchemeAndHost(c *gin.Context) (string, string) {
	host := c.Request.Host
	if forwardHost := c.GetHeader("X-Forwarded-Host"); forwardHost != "" {
		host = forwardHost
	}
	scheme := "https"
	if strings.Contains(host, "localhost") {
		scheme = "http"
	}
	return scheme, host
}

type UserBindByHeader struct {
	Id       string `header:"X-User-Id"`
	User     string `header:"X-User-Account"`
	Email    string `header:"X-User-Email"`
	Name     string `header:"X-User-Name"`
	Language string `header:"X-User-Language"`
}

func (t ssoRepository) CallbackAndSaveUser(c *gin.Context) {
	if c.Query("state") == "" {
		helpers.ResponseBadRequestError(c, "missing state")
		return
	}
	mysession := sessions.Default(c)
	stateObj := mysession.Get("state")
	if stateObj == nil || (stateObj.(string) != c.Query("state")) {
		helpers.ResponseBadRequestError(c, "invalid state")
		return
	}

	var user UserBindByHeader
	err := c.BindHeader(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var findUser models.Users
	err = config.DB.Collection("Users").FindOne(context.TODO(), bson.D{{Key: "users_source_id", Value: user.Id}}).Decode(&findUser)

	if err == nil && findUser.UsersId.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		return
	}

	_, err = saveUserInfo(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	return_to := mysession.Get("return_to")
	mysession.Clear()
	if return_to != nil {
		c.Redirect(http.StatusSeeOther, return_to.(string))
		return
	}
	mysession.Save()

}