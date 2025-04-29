package repository

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"oosa/internal/config"
	"oosa/internal/helpers"
	"oosa/internal/models"
	"oosa/internal/structs"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewSSoRepository(url *url.URL) ssoRepository {
	return ssoRepository{
		ssoRegisterUrl: url,
	}
}

type ssoRepository struct {
	ssoRegisterUrl *url.URL
}

var defaultFriendAutoAdd = 0
var defaultFriendTakeMestatus = false

// TODO: add user db
func saveUserInfo(c context.Context, user *structs.UserBindByHeader) (*models.Users, error) {
	var User models.Users

	insert := models.Users{
		UsersSource:                           3,
		UsersSourceId:                         user.Id,
		UsersName:                             user.Name,
		UsersEmail:                            user.Email,
		UsersUsername:                         user.User,
		UsersObject:                           user.User,
		UsersAvatar:                           user.Avatar,
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
	result, _ := config.DB.Collection("Users").InsertOne(c, insert)
	config.DB.Collection("Users").FindOne(c, bson.D{{Key: "_id", Value: result.InsertedID}}).Decode(&User)
	return &User, nil
}

func (t ssoRepository) Register(c *gin.Context) {
	if c.Query("return_to") == "" {
		helpers.ResponseBadRequestError(c, "missing return_to")
		return
	}

	t.ssoRegisterUrl.RawQuery = fmt.Sprintf("return_to=%s", url.QueryEscape(c.Query("return_to")))
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

func (t ssoRepository) CallbackAndSaveUser(c *gin.Context) {
	mysession := sessions.Default(c)
	defer func() {
		return_to := mysession.Get("return_to")
		mysession.Clear()
		if return_to != nil {
			c.Redirect(http.StatusSeeOther, return_to.(string))
			return
		}
		mysession.Save()
	}()
	if c.Query("state") == "" {
		return
	}

	stateObj := mysession.Get("state")
	if stateObj == nil || (stateObj.(string) != c.Query("state")) {
		return
	}

	var user structs.UserBindByHeader
	err := c.BindHeader(&user)
	if err != nil {
		return
	}

	var findUser models.Users
	err = config.DB.Collection("Users").FindOne(c, bson.D{{Key: "users_source_id", Value: user.Id}}).Decode(&findUser)
	if err != mongo.ErrNoDocuments || !findUser.UsersId.IsZero() {
		return
	}

	_, err = saveUserInfo(c, &user)
	if err != nil {
		return
	}

}
