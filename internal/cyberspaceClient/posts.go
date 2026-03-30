package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	//client "github.com/johannesalke/CyberspaceTUI/internal/cyberspaceClient"
	//"net/http"
	"os"
	"os/exec"
	"time"
)

type GetPostsResponse struct {
	Data   []Post `json:"data"`
	Cursor string `json:"cursor"`
}

type OnePostResponse struct {
	Data Post `json:"data"`
}

type Post struct {
	PostID         string    `json:"postId"`
	AuthorID       string    `json:"authorId"`
	AuthorUsername string    `json:"authorUsername"`
	Content        string    `json:"content"`
	Topics         []string  `json:"topics"`
	RepliesCount   int       `json:"repliesCount"`
	BookmarksCount int       `json:"bookmarksCount"`
	IsPublic       bool      `json:"isPublic"`
	IsNSFW         bool      `json:"isNSFW"`
	Attachments    any       `json:"attachments"`
	CreatedAt      time.Time `json:"createdAt"`
	Deleted        bool      `json:"deleted"`
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

func (c *APIClient) GetPosts(limit int, cursor string) (posts []Post, newCursor string, err error) {
	url := makeGetUrl(c.ApiUrl+"/posts", limit, cursor)

	req, err := makeRequest("GET", url, c.Tokens, nil)
	if err != nil {
		return []Post{}, cursor, fmt.Errorf("Error forming request: %s", err)
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return nil, cursor, fmt.Errorf("Error retrieving Posts: %s", err)
	}

	var getPostsResponse GetPostsResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&getPostsResponse)
	if err != nil {
		panic(err)
	}
	//fmt.Print(getNotificationsReply)
	c.Cursors["posts_standard"] = getPostsResponse.Cursor
	return getPostsResponse.Data, getPostsResponse.Cursor, nil

}

func (c *APIClient) GetPostById(post_id string) (Post, error) {

	req, err := makeRequest("GET", "https://api.cyberspace.online/v1/posts/"+post_id, c.Tokens, nil)
	if err != nil {
		return Post{}, fmt.Errorf("Error forming request: %s", err)
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return Post{}, fmt.Errorf("Error requesting post by ID: %s", err)
	}
	var postConfirm OnePostResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&postConfirm)
	if err != nil {
		panic(err)
	}
	fmt.Print(postConfirm)
	return postConfirm.Data, nil
}

func (c *APIClient) CreatePost(tokens AuthTokens) error {

	content := WritePost()
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
	req, err := makeRequest("POST", c.ApiUrl+"/posts", tokens, bytes.NewBuffer(postJson))
	if err != nil {
		return fmt.Errorf("Error making post request:%s", err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("Error sending post request:%s", err)
	}

	var postConfirm OnePostResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&postConfirm)
	if err != nil {
		return fmt.Errorf("Error decoding post json:%s", err)
	}
	fmt.Print(postConfirm)
	return nil
}

func WritePost() string {
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

	fmt.Println("\nMessage:")
	fmt.Println(string(content))
	return string(content)
}

func (c *APIClient) DeletePost(postId string) error {

	req, err := makeRequest("DELETE", c.ApiUrl+"/posts/"+postId, c.Tokens, nil)
	if err != nil {
		return fmt.Errorf("Error forming delete request: %s", err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("Error during the request process: %s", err)
	}

	if res.StatusCode == 200 { //Check result based on response code.
		fmt.Printf("The post was successfully deleted.\n")
	} else if res.StatusCode == 404 {
		fmt.Printf("No post with that id found.\n")
	} else if res.StatusCode == 403 {
		fmt.Printf("You do not have authority to delete this post.\n")
	} else {
		fmt.Printf("Something went wrong.\n")
	}
	return nil
}
