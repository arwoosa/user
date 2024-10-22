package repository

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

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
	registerUrl := *c.Request.URL
	registerUrl.Scheme, registerUrl.Host = getSchemeAndHost(c)
	state := uuid.NewString()
	mysession := sessions.Default(c)
	mysession.Set("state", state)
	registerUrl.Path = fmt.Sprintf("/api%s/finish", registerUrl.Path)
	registerUrl.RawQuery = fmt.Sprintf("state=%s", state)

	ssoRegisterUrl.RawQuery = fmt.Sprintf("return_to=%s", url.QueryEscape(registerUrl.String()))
	mysession.Set("return_to", c.Query("return_to"))
	mysession.Save()
	c.Redirect(http.StatusSeeOther, ssoRegisterUrl.String())
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
	Id       string `header:"X-Userid"`
	User     string `header:"X-User"`
	Email    string `header:"X-Email"`
	Name     string `header:"X-Name"`
	Language string `header:"X-Language"`
}

func (t SsoRepository) CallbackAndSaveUser(c *gin.Context) {
	mysession := sessions.Default(c)
	stateObj := mysession.Get("state")
	if stateObj != nil && stateObj.(string) != c.Query("state") {
		t.Register(c)
		return
	}
	var user UserBindByHeader
	err := c.BindHeader(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	return_to := mysession.Get("return_to")
	mysession.Clear()
	if return_to != nil {
		fmt.Println(return_to)
		c.Redirect(http.StatusSeeOther, return_to.(string))
		return
	}
	mysession.Save()
}
