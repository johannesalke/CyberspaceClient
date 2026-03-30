package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type CreateReplyInput struct {
	PostID        string `json:"postId"`
	Content       string `json:"content"`
	ParentReplyID string `json:"parentReplyId"`
}

type getRepliesResponse struct {
	Data   []Reply `json:"data"`
	Cursor string  `json:"cursor"`
}

type oneReplyResponse struct {
	Data Reply `json:"data"`
}

type Reply struct {
	PostID            string    `json:"postId"`
	ParentPostAuthor  string    `json:"parentPostAuthor"`
	AuthorID          string    `json:"authorId"`
	AuthorUsername    string    `json:"authorUsername"`
	Content           string    `json:"content"`
	Deleted           bool      `json:"deleted"`
	SavesCount        int       `json:"savesCount"`
	CreatedAt         time.Time `json:"createdAt"`
	ReplyID           string    `json:"replyId"`
	ParentReplyID     string    `json:"parentReplyId,omitempty"`
	ParentReplyAuthor string    `json:"parentReplyAuthor,omitempty"`
	Attachments       []struct {
		Type string `json:"type"`
		Src  string `json:"src"`
	} `json:"attachments,omitempty"`
	HasImageAttachment bool `json:"hasImageAttachment,omitempty"`
}

func (c *APIClient) GetReplies(postID string, limit int, cursor string) (replies []Reply, newCursor string, err error) {
	url := makeGetUrl(c.ApiUrl+"/posts/"+postID+"/replies", limit, cursor)

	req, err := makeRequest("GET", url, c.Tokens, nil)
	if err != nil {
		return []Post{}, cursor, fmt.Errorf("Error forming request: %s", err)
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return nil, cursor, fmt.Errorf("Error retrieving Posts: %s", err)
	}

	var getPostsResponse getRepliesResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&getPostsResponse)
	if err != nil {
		panic(err)
	}
	//fmt.Print(getNotificationsReply)
	cursor_key := "replies_" + postID
	c.Cursors[cursor_key] = getPostsResponse.Cursor

	return getPostsResponse.Data, getPostsResponse.Cursor, nil

}

func (c *APIClient) CreateReply(tokens AuthTokens, replyInput CreateReplyInput) error {

	//content := WritePost()

	replyJson, err := json.Marshal(replyInput)
	if err != nil {
		panic(err)
	}
	req, err := makeRequest("POST", c.ApiUrl+"/replies", tokens, bytes.NewBuffer(replyJson))
	if err != nil {
		return fmt.Errorf("Error making reply request:%s", err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("Error sending reply request:%s", err)
	}

	var postConfirm OnePostResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&postConfirm)
	if err != nil {
		return fmt.Errorf("Error decoding reply json:%s", err)
	}
	fmt.Print(postConfirm)
	return nil
}

func (c *APIClient) DeleteReply(replyID string) error {

	req, err := makeRequest("DELETE", c.ApiUrl+"/replies/"+replyID, c.Tokens, nil)
	if err != nil {
		return fmt.Errorf("Error forming delete request: %s", err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("Error during the request process: %s", err)
	}

	if res.StatusCode == 200 { //Check result based on response code.
		fmt.Printf("The reply was successfully deleted.\n")
	} else if res.StatusCode == 404 {
		fmt.Printf("No reply with that id found.\n")
	} else if res.StatusCode == 403 {
		fmt.Printf("You do not have authority to delete this reply.\n")
	} else {
		fmt.Printf("Something went wrong.\n")
	}
	return nil
}
