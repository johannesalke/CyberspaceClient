package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendRequestRefreshesTokenOn401(t *testing.T) {
	var requests []string
	var authHeaders []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Path)
		authHeaders = append(authHeaders, r.Header.Get("Authorization"))
		switch r.URL.Path {
		case "/auth/refresh":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":{"idToken":"new-token","rtbdToken":"new-rtdb"}}`))
		default:
			if len(requests) <= 2 {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"expired"}}`))
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"data":[{"postId":"1"}],"cursor":"next"}`))
			}
		}
	}))
	defer server.Close()

	c := InitAPIClient()
	c.ApiUrl = server.URL
	c.Tokens = AuthTokens{IDToken: "old-token", RefreshToken: "refresh-token"}

	req, err := makeRequest("GET", server.URL+"/posts", c.Tokens, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		t.Fatalf("sendRequest: %v", err)
	}
	res.Body.Close()

	if len(requests) != 3 {
		t.Fatalf("expected 3 requests, got %d: %v", len(requests), requests)
	}
	if requests[0] != "/posts" || requests[1] != "/auth/refresh" || requests[2] != "/posts" {
		t.Fatalf("unexpected request sequence: %v", requests)
	}
	if !strings.HasPrefix(authHeaders[2], "Bearer new-token") {
		t.Fatalf("retry used %q, want Bearer new-token", authHeaders[2])
	}
	if c.Tokens.IDToken != "new-token" {
		t.Fatalf("client token = %q, want new-token", c.Tokens.IDToken)
	}
}

func TestSendRequestDoesNotRetryOnNon401(t *testing.T) {
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := InitAPIClient()
	c.ApiUrl = server.URL
	c.Tokens = AuthTokens{IDToken: "t", RefreshToken: "r"}

	req, err := makeRequest("GET", server.URL+"/posts", c.Tokens, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		t.Fatalf("sendRequest: %v", err)
	}
	res.Body.Close()

	if count != 1 {
		t.Fatalf("expected 1 request, got %d", count)
	}
}

func TestSendRequestWithoutRefreshTokenDoesNotRetry(t *testing.T) {
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	c := InitAPIClient()
	c.ApiUrl = server.URL
	c.Tokens = AuthTokens{IDToken: "t"}

	req, err := makeRequest("GET", server.URL+"/posts", c.Tokens, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		t.Fatalf("sendRequest: %v", err)
	}
	res.Body.Close()

	if count != 1 {
		t.Fatalf("expected 1 request, got %d", count)
	}
}

func TestExpectSuccessParsesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":{"code":"FORBIDDEN","message":"nope"}}`))
	}))
	defer server.Close()

	c := InitAPIClient()
	c.ApiUrl = server.URL
	c.Tokens = AuthTokens{IDToken: "t"}

	req, err := makeRequest("GET", server.URL+"/posts", c.Tokens, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		t.Fatalf("sendRequest: %v", err)
	}
	err = c.expectSuccess(res, "retrieving posts")
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
	if !strings.Contains(err.Error(), "FORBIDDEN") || !strings.Contains(err.Error(), "nope") {
		t.Fatalf("error did not carry API details: %v", err)
	}
}
