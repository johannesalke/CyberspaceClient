package client

import (
	"net/http"
)

type APIClient struct {
	Client    *http.Client
	Tokens    AuthTokens
	UserID    string
	Username  string
	PostCache map[string]Post
	ApiUrl    string
}
