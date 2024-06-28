package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"oosa/internal/auth"
	"oosa/internal/config"
	"oosa/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthOOSA struct {
	User  models.Users `json:"user"`
	Token string       `json:"token"`
}

type AuthLineRequest struct {
	Code  string `form:"code" binding:"required"`
	State string `form:"state" binding:"required"`
}

type LineAccessTokenRequest struct {
	GrantType    string `form:"grant_type" binding:"required"`
	Code         string `form:"code" binding:"required"`
	RedirectURI  string `form:"redirect_uri" binding:"required"`
	ClientID     string `form:"client_id" binding:"required"`
	ClientSecret string `form:"client_secret" binding:"required"`
}

type LineAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

type LineUserInfoResponse struct {
	UserID  string `json:"sub"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type AuthFacebookRequest struct {
	Id   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type AuthEmailRequest struct {
	Name       string `json:"name" validate:"required"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	IsBusiness bool   `json:"is_business"`
}

func AuthenticateUser(c *gin.Context, user models.Users) {
	token, err := auth.GenerateJWT(user)
	if err != nil {
		ResponseNoData(c, err.Error())
		return
	}

	auth := AuthOOSA{
		User:  user,
		Token: token,
	}
	c.JSON(200, auth)
}

func GetLineAccessToken(params AuthLineRequest) (*LineAccessTokenResponse, error) {
	clientID := os.Getenv("OAUTH_LINE_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_LINE_CLIENT_SECRET")

	tokenRequest := LineAccessTokenRequest{
		GrantType:    "authorization_code",
		Code:         params.Code,
		RedirectURI:  config.APP.OauthLineRedirect,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	accessToken, err := makeLineAccessTokenRequest(tokenRequest)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func makeLineAccessTokenRequest(request LineAccessTokenRequest) (*LineAccessTokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", request.GrantType)
	data.Set("code", request.Code)
	data.Set("redirect_uri", request.RedirectURI)
	data.Set("client_id", request.ClientID)
	data.Set("client_secret", request.ClientSecret)

	resp, err := http.PostForm("https://api.line.me/oauth2/v2.1/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get access token: %s", resp.Status)
	}

	var result LineAccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func GetUserInfo(accessToken string) (*LineUserInfoResponse, error) {
	url := "https://api.line.me/oauth2/v2.1/userinfo"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	var userInfo LineUserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPassword(password string, match string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(match))
	if err != nil {
		return false
	}

	return true
}

func GetAuthUser(c *gin.Context) models.Users {
	user, _ := c.Get("user")
	userDetail := user.(*models.Users)

	return *userDetail
}
