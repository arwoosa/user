package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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

// TODO: add login
func saveUserInfo(user *UserBindByHeader) error {
	fmt.Println(user)
	return nil
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

	for key, value := range c.Request.Header {
		fmt.Printf("%s: %s\n", key, value)
	}

	var user UserBindByHeader
	err := c.BindHeader(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = saveUserInfo(&user)
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

func (SsoRepository) ResponseHeader(c *gin.Context) {
	c.JSON(http.StatusOK, c.Request.Header)
}

type flowResponse struct {
	FlowId string `json:"id"`
}

func getSsoLoginFlowId() (string, error) {
	req, err := http.NewRequest("GET", "http://kratos:4433/self-service/login/api", nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var flowResponse flowResponse

	err = json.NewDecoder(resp.Body).Decode(&flowResponse)
	if err != nil {
		return "", err
	}
	for k, v := range resp.Header {
		fmt.Printf("%s: %s\n", k, v)
	}

	return flowResponse.FlowId, nil
}

func (SsoRepository) LoginByEmail(c *gin.Context) {
	flowid, err := getSsoLoginFlowId()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jsonString := `{"identifier": "AlexDavis344@test.com", "password": "arwork8888", "method": "password"}`
	req, err := http.NewRequest(
		"POST",
		"http://kratos:4433/self-service/login?flow="+flowid,
		bytes.NewReader([]byte(jsonString)))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println(resp.StatusCode, string(data))

	c.JSON(http.StatusOK, gin.H{"message": "Login"})
}

type ssoProviderResponse struct {
	RedirectUrl string `json:"redirect_browser_to"`
}

func (SsoRepository) LoginByProvider(c *gin.Context) {
	provider := c.Param("provider")
	flowid, err := getSsoLoginFlowId()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jsonString := fmt.Sprintf(`{"provider": "%s"}`, provider)

	req, err := http.NewRequest(
		"POST",
		"http://kratos:4433/self-service/login?flow="+flowid,
		bytes.NewReader([]byte(jsonString)))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider not found"})
		return
	} else if resp.StatusCode == http.StatusUnprocessableEntity {
		fmt.Println("login by oidc")
		for k, v := range resp.Header {
			if k != "Set-Cookie" {
				continue
			}
			for _, vv := range v {
				c.Writer.Header().Add(k, vv)
			}
		}

		var providerResp ssoProviderResponse
		json.NewDecoder(resp.Body).Decode(&providerResp)
		fmt.Println(providerResp.RedirectUrl)
		loginUrl, err := url.Parse(providerResp.RedirectUrl)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		switch provider {
		case "line":
			queryValues := loginUrl.Query()
			queryValues.Del("redirect_uri")
			schema, host := getSchemeAndHost(c)
			redirect := fmt.Sprintf("%s://%s/api/sso/login/%s/callback", schema, host, provider)
			queryValues.Add("redirect_uri", redirect)
			loginUrl.RawQuery = queryValues.Encode()
		}

		c.Header("Location", providerResp.RedirectUrl)
		c.JSON(http.StatusOK, gin.H{"redirect": loginUrl.String()})
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(resp.StatusCode, string(data))

	c.JSON(http.StatusOK, gin.H{"message": "Login"})
}

type oidcResponse struct {
	Id string `json:"id"`
	UI struct {
		Action string  `json:"action"`
		Nodes  []*node `json:"nodes"`
	} `json:"ui"`
}

type node struct {
	Type       string
	Attributes struct {
		Name  string
		Value string
	} `json:"attributes"`
}

func (SsoRepository) ProviderCallback(c *gin.Context) {
	provider := c.Param("provider")

	myurl := &url.URL{
		Scheme: "http",
		Host:   "kratos:4433",
		Path:   fmt.Sprintf("/self-service/methods/oidc/callback/%s", provider),
	}
	myurl.RawQuery = c.Request.URL.RawQuery
	req, err := http.NewRequest(
		"GET",
		myurl.String(),
		nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for k, v := range c.Request.Header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		var ocidResp oidcResponse
		json.NewDecoder(resp.Body).Decode(&ocidResp)
		result := gin.H{"state": "unregister", "id": ocidResp.Id}

		for k, v := range resp.Header {
			for _, vv := range v {
				c.Header(k, vv)
			}
		}
		for _, node := range ocidResp.UI.Nodes {
			if node.Type != "input" {
				continue
			}
			result[node.Attributes.Name] = node.Attributes.Value
		}
		// 尚末註冊
		c.JSON(http.StatusOK, result)
		return
	} else if resp.StatusCode == http.StatusOK {
		var successResp ocidRegisterSuccessResp
		json.NewDecoder(resp.Body).Decode(&successResp)
		saveUserInfo(&UserBindByHeader{
			Id:       successResp.Session.Identity.Id,
			Email:    successResp.Session.Identity.Traits.Email,
			Name:     successResp.Session.Identity.Traits.Name,
			User:     successResp.Session.Identity.Traits.Username,
			Language: successResp.Session.Identity.Traits.Language,
		})
		c.JSON(http.StatusOK, gin.H{"token": successResp.Token})
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(resp.StatusCode, string(data))
	// 儲存使用者在oosa

	c.Writer.Write(data)
}

type ocidRegisterSuccessResp struct {
	Token   string `json:"session_token"`
	Session struct {
		Identity struct {
			Id     string `json:"id"`
			Traits struct {
				Email    string `json:"email"`
				Language string `json:"language"`
				Name     string `json:"name"`
				Picture  string `json:"picture"`
				Username string `json:"username"`
			} `json:"traits"`
		} `json:"identity"`
	} `json:"session"`
}

func (SsoRepository) OcidRegister(c *gin.Context) {
	fmt.Println("ocid register")
	var data map[string]any

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(data)

	postUrl := "http://kratos:4433/self-service/registration?flow=" + data["id"].(string)
	fmt.Println(postUrl)
	delete(data, "id")
	delete(data, "state")
	postData, _ := json.Marshal(data)
	fmt.Println(string(postData))
	req, err := http.NewRequest(
		"POST",
		postUrl,
		bytes.NewReader(postData))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for k, v := range c.Request.Header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(resp.StatusCode, string(respData))

	c.Writer.Write(respData)
}

func (SsoRepository) Recover(c *gin.Context) {

	req, err := http.NewRequest("GET", "http://kratos:4433/self-service/recovery/api", nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	var flowResponse flowResponse
	err = json.NewDecoder(resp.Body).Decode(&flowResponse)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(resp.StatusCode, flowResponse.FlowId)

	jsonString := fmt.Sprintf(`{"email": "%s", "method": "code"}`, "AlexDavis344@test.com")

	req, err = http.NewRequest(
		"POST",
		"http://kratos:4433/self-service/recovery?flow="+flowResponse.FlowId,
		bytes.NewReader([]byte(jsonString)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(resp.StatusCode, string(data))
	c.JSON(http.StatusOK, gin.H{"message": "Recover"})

}

func (SsoRepository) RecoverCode(c *gin.Context) {
	c.Query("flow")
}

func (SsoRepository) Error(c *gin.Context) {
	id := c.Query("id")
	fmt.Println("http://kratos:4433/self-service/error?id=" + id)
	req, err := http.NewRequest(
		"GET",
		"http://kratos:4433/self-service/errors?id="+id, nil)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Writer.Write(data)
}

type Form struct {
	Name string                `form:"name" binding:"required"`
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func (SsoRepository) PutTest(c *gin.Context) {
	f := Form{}
	err := c.ShouldBind(&f)
	fmt.Println(err, f)
}
