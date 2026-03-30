package client

import (
	"net/http"
)

type APIClient struct {
	Client            *http.Client
	Tokens            AuthTokens
	ApiUrl            string
	UserID            string
	Username          string
	PostCache         map[string]Post         // key:PostID
	NotificationCache map[string]Notification // key:PostID
	Cursors           map[string]string       // key: whatever you want
	LastStatusCode    int
}

const CyberspaceApiUrl = "https://api.cyberspace.online/v1"

func InitAPIClient() APIClient {
	return APIClient{
		ApiUrl:            CyberspaceApiUrl,
		Client:            &http.Client{},
		PostCache:         make(map[string]Post),
		NotificationCache: make(map[string]Notification),
		Cursors:           make(map[string]string),
	}
}
