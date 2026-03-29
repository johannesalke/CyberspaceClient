package client

import (
	//"github.com/johannesalke/CyberspaceTUI/internal/auth"
	"fmt"
	"io"
	"net/http"
)

func makeRequest(method, url string, tokens AuthTokens, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+tokens.IDToken)

	return req, nil
}

func makeGetUrl(url string, limit int, cursor string) string {
	if limit == 0 {
		limit = 20
	}

	url += fmt.Sprintf("?limit=%s", limit)
	if cursor != "" {
		url += fmt.Sprintf("&cursor=%s", cursor)
	}
	return url
}

func (c APIClient) sendRequest(req *http.Request) (*http.Response, error) {
	res, err := c.Client.Do(req)
	c.LastStatusCode = res.StatusCode
	if err != nil {
		return res, fmt.Errorf("Error sending request: %s", err)
	}
	return res, nil
}
