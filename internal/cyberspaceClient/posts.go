package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	//client "github.com/johannesalke/CyberspaceTUI/internal/cyberspaceClient"
	//"net/http"
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

type CreatePostConfirmation struct {
	Data struct {
		PostID string `json:"postId"`
	} `json:"data"`
}

func (c *APIClient) GetPosts(limit int, cursor string) (posts []Post, newCursor string, err error) {
	page, next, err := c.fetchFeedPage(limit, cursor)
	if err != nil {
		return page, cursor, err
	}
	for _, post := range page {
		c.PostCache[post.PostID] = post
	}
	c.Cursors["feed"] = next
	return page, next, nil
}

// fetchFeedPage retrieves a single page of the feed without touching the
// pagination cursor or the post cache. Shared by the feed pager and the
// jukebox's deeper track scan.
func (c *APIClient) fetchFeedPage(limit int, cursor string) ([]Post, string, error) {
	url := makeGetUrl(c.ApiUrl+"/posts", limit, cursor)

	req, err := makeRequest("GET", url, c.Tokens, nil)
	if err != nil {
		return []Post{}, cursor, fmt.Errorf("Error forming request: %s", err)
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return nil, cursor, fmt.Errorf("Error retrieving Posts: %s", err)
	}
	if err := c.expectSuccess(res, "retrieving posts"); err != nil {
		return nil, cursor, err
	}

	var getPostsResponse GetPostsResponse
	if err := json.NewDecoder(res.Body).Decode(&getPostsResponse); err != nil {
		return nil, cursor, fmt.Errorf("error decoding posts: %s", err)
	}
	return getPostsResponse.Data, getPostsResponse.Cursor, nil
}

// GetPostsForTracks walks the feed from startCursor until it has collected up
// to limit posts (or the feed runs out), for the jukebox to build its
// catalogue from. It returns the cursor to resume from so the scan can
// continue deeper in later calls; cursor is empty when the feed is exhausted.
// Unlike GetPosts it never advances the feed page cursor, so browsing the
// feed and scanning for tracks remain independent.
func (c *APIClient) GetPostsForTracks(limit int, startCursor string) ([]Post, string, error) {
	posts := make([]Post, 0, limit)
	cursor := startCursor
	for len(posts) < limit {
		page, next, err := c.fetchFeedPage(min(limit-len(posts), 50), cursor)
		if err != nil {
			return posts, cursor, err
		}
		for _, post := range page {
			c.PostCache[post.PostID] = post
		}
		posts = append(posts, page...)
		if next == "" || next == cursor || len(page) == 0 {
			return posts, "", nil
		}
		cursor = next
	}
	return posts, cursor, nil
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
	if err := c.expectSuccess(res, "requesting post"); err != nil {
		return Post{}, err
	}
	var postConfirm OnePostResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&postConfirm)
	if err != nil {
		return Post{}, fmt.Errorf("error decoding post: %s", err)
	}
	c.PostCache[postConfirm.Data.PostID] = postConfirm.Data
	//fmt.Print(postConfirm)
	return postConfirm.Data, nil
}

func (c *APIClient) CreatePost(postInput CreatePostInput) (Post, error) {
	writeInCLI := postInput.Content == "" //Check if the contents of the post have been handed in via argument. If not, use terminal text editor to write post.
	if writeInCLI {
		content := WriteContent()         //See: utilities
		topics := WriteTopics([]string{}) //See: utilities
		postInput = CreatePostInput{
			Content:  content,
			Topics:   topics,
			IsPublic: false,
			IsNSFW:   false,
		}
	}
	if writeInCLI {
		if ConfirmPostIntention() == false {
			return Post{}, nil
		}
	}
	postJson, err := json.Marshal(postInput)
	if err != nil {
		return Post{}, fmt.Errorf("error encoding post: %s", err)
	}
	req, err := makeRequest("POST", c.ApiUrl+"/posts", c.Tokens, bytes.NewBuffer(postJson))
	if err != nil {
		return Post{}, fmt.Errorf("Error making post request:%s", err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return Post{}, fmt.Errorf("Error sending post request:%s", err)
	}
	if err := c.expectSuccess(res, "posting"); err != nil {
		return Post{}, err
	}

	var postConfirm CreatePostConfirmation

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&postConfirm)
	if err != nil {
		return Post{}, fmt.Errorf("Error decoding post json:%s", err)
	}
	//fmt.Print(postConfirm)
	postMade := Post{ //The response is just a post ID, so it's necessary to manually create the Post object for rendering on the client side. The alternative is to request the post from the server, but for optimization reasons (and becasue that is not possible for replies) I'm doing it the direct way.
		Content: postInput.Content, Topics: postInput.Topics, PostID: postConfirm.Data.PostID,
		IsPublic: postInput.IsPublic, IsNSFW: postInput.IsNSFW, AuthorUsername: c.Username,
	}
	c.PostCache[postMade.PostID] = postMade
	return postMade, nil
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
		//fmt.Printf("The post was successfully deleted.\n")
	} else if res.StatusCode == 404 {
		return fmt.Errorf("No post with that id found.\n")
	} else if res.StatusCode == 403 {
		return fmt.Errorf("You do not have authority to delete this post.\n")
	} else {
		return fmt.Errorf("Something went wrong.\n")
	}
	return nil
}
