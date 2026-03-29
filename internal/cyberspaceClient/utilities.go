package client

import (
	//"github.com/johannesalke/CyberspaceTUI/internal/auth"
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
