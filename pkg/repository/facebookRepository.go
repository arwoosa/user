package repository

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"oosa/internal/config"

	"github.com/gin-gonic/gin"
)

type FacebookOauthResponse struct {
	Errors   []string         `json:"errors"`
	Messages []string         `json:"messages"`
	Result   CloudflareResult `json:"result"`
	Success  bool             `json:"success"`
}

type FacebookOauthResult struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type FacebookRepository struct{}

func (r FacebookRepository) Retrieve(c *gin.Context, accessToken string) FacebookOauthResult {
	var response FacebookOauthResult
	endpoint := config.APP.FacebookUrl + "me?access_token=" + accessToken
	client := &http.Client{}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatal(err)
		return response
	}

	// Create Response
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return response
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal(body, &response)

	return response
}
