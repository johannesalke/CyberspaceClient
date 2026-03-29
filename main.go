package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/johannesalke/CyberspaceTUI/internal/auth"
)

type Config struct {
	apiUrl string
	cache  map[string]any
	tokens auth.AuthTokens
	client http.Client
}

func main() {
	cfg := Config{apiUrl: "https://api.cyberspace.online/v1"}
	//client := http.NewClientHandler()
	cfg.tokens = auth.Login(cfg.apiUrl)
	fmt.Printf("authToken: %.10s", cfg.tokens.IDToken)

	scanner := bufio.NewScanner(os.Stdin)
	CreatePost(cfg.apiUrl, cfg.tokens)
	for true {
		scanner.Scan()
		input := scanner.Text()
		args := strings.Split(input, " ")
		if len(args) == 0 {
			continue
		}

		//cmd := args[0]

	}

}

////////////////////////////| Posts |///////////////////////////

type GetPostsResponse struct {
	Data   []Post `json:"data"`
	Cursor string `json:"cursor"`
}

type OnePostResponse struct {
	Data Post `json:"data"`
}

type Post struct {
	PostID         string        `json:"postId"`
	AuthorID       string        `json:"authorId"`
	AuthorUsername string        `json:"authorUsername"`
	Content        string        `json:"content"`
	Topics         []string      `json:"topics"`
	RepliesCount   int           `json:"repliesCount"`
	BookmarksCount int           `json:"bookmarksCount"`
	IsPublic       bool          `json:"isPublic"`
	IsNSFW         bool          `json:"isNSFW"`
	Attachments    []interface{} `json:"attachments"`
	CreatedAt      time.Time     `json:"createdAt"`
	Deleted        bool          `json:"deleted"`
}

type CreatePostInput struct {
	Content     string   `json:"content"`
	Topics      []string `json:"topics"`
	IsPublic    bool     `json:"isPublic"`
	IsNSFW      bool     `json:"isNSFW"`
	Attachments []struct {
		Type   string `json:"type"`
		Src    string `json:"src"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"attachments"`
}

func CreatePost(url string, tokens auth.AuthTokens) error {

	content := EditPost()
	postInput := CreatePostInput{
		Content:  content,
		Topics:   []string{"test", "api", "cli"},
		IsPublic: true,
		IsNSFW:   false,
	}
	postJson, err := json.Marshal(postInput)
	if err != nil {
		panic(err)
	}
	req, err := makeRequest("POST", url+"/posts", tokens, bytes.NewBuffer(postJson))
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	var postConfirm OnePostResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&postConfirm)
	if err != nil {
		panic(err)
	}
	fmt.Print(postConfirm)
	return nil
}

func EditPost() string {
	tmpFile, err := os.CreateTemp("", "message-*.txt")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // fallback
	}

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		panic(err)
	}

	fmt.Println("Message:")
	fmt.Println(string(content))
	return string(content)
}

/*
func (cfg *Config) sendRequest() {
	body := []byte(`{"name":"John"}`)

	req, err := makeRequest()
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer ") //+cfg.tokens.IDToken

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func

	res, err := http.DefaultClient.Do(req)

	if res.StatusCode == 401 {
		tokens = auth.TokenRefresh(url, tokens)
		req, err := http.NewRequest(method, url, body)
		req.Header.Set("Authorization", "Bearer "+tokens.IDToken)

		res, err = http.DefaultClient.Do(req)

		if res.StatusCode
	}
	return req, nil
*/
