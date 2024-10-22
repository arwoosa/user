package repository

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SsoRepository struct{}

var ssoRegisterUrl *url.URL

func init() {
	var err error
	registerUrl := os.Getenv("SSO_REGISTER_URL")
	if registerUrl == "" {
		panic("SSO_REGISTER_URL not set")
	}
	ssoRegisterUrl, err = url.Parse(registerUrl)
	if err != nil {
		panic("failed to parse SSO_REGISTER_URL: " + err.Error())
	}
}

func (t SsoRepository) Register(c *gin.Context) {
	mysession := sessions.Default(c)

	stateObj := mysession.Get("state")
	if stateObj != nil && stateObj.(string) == c.Query("state") {
		t.callbackAndSaveUser(c)
		return
	}

	registerUrl := *c.Request.URL
	registerUrl.Scheme = "https"
	registerUrl.Host = c.Request.Host
	state := uuid.NewString()
	mysession.Set("state", state)
	registerUrl.Path = registerUrl.Path + "/finish"
	registerUrl.RawQuery = fmt.Sprintf("state=%s", state)

	ssoRegisterUrl.RawQuery = fmt.Sprintf("return_to=%s", url.QueryEscape(registerUrl.String()))
	mysession.Set("return_to", c.Query("return_to"))
	mysession.Save()

	c.Redirect(http.StatusSeeOther, ssoRegisterUrl.String())
}

type UserBindByHeader struct {
	Id       string `header:"X-Userid"`
	User     string `header:"X-User"`
	Email    string `header:"X-Email"`
	Name     string `header:"X-Name"`
	Language string `header:"X-Language"`
}

func (t SsoRepository) callbackAndSaveUser(c *gin.Context) {
	var user UserBindByHeader
	err := c.BindHeader(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mysession := sessions.Default(c)
	return_to := mysession.Get("return_to")
	mysession.Clear()
	if return_to != nil {
		c.Redirect(http.StatusSeeOther, return_to.(string))
		return
	}
	mysession.Save()
}
