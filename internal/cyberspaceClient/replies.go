package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

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

type CreateReplyInput struct {
	PostID        string `json:"postId"`
	Content       string `json:"content"`
	ParentReplyID string `json:"parentReplyId"`
}
type createReplyResponse struct {
	Data struct {
		ReplyID string `json:"replyId"`
	} `json:"data"`
}

func (c *APIClient) GetReplies(postID string, limit int, cursor string) (replies []Reply, newCursor string, err error) {
	url := makeGetUrl(c.ApiUrl+"/posts/"+postID+"/replies", limit, cursor)

	req, err := makeRequest("GET", url, c.Tokens, nil)
	if err != nil {
		return nil, cursor, fmt.Errorf("Error forming request: %s", err)
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return nil, cursor, fmt.Errorf("Error retrieving Posts: %s", err)
	}

	var getRepliesResponse getRepliesResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&getRepliesResponse)
	if err != nil {
		panic(err)
	}
	//fmt.Print(getNotificationsReply)
	cursor_key := "replies_" + postID
	c.Cursors[cursor_key] = getRepliesResponse.Cursor
	for _, reply := range getRepliesResponse.Data {
		c.ReplyCache[reply.ReplyID] = reply
	}

	return getRepliesResponse.Data, getRepliesResponse.Cursor, nil

}

func (c *APIClient) GetReplyById(reply_id string) (Reply, error) {

	req, err := makeRequest("GET", "https://api.cyberspace.online/v1/replies/"+reply_id, c.Tokens, nil)
	if err != nil {
		return Reply{}, fmt.Errorf("Error forming request: %s", err)
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return Reply{}, fmt.Errorf("Error requesting post by ID: %s", err)
	}
	var oneReply oneReplyResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&oneReply)
	if err != nil {
		panic(err)
	}
	c.ReplyCache[oneReply.Data.ReplyID] = oneReply.Data
	//fmt.Print(postConfirm)
	return oneReply.Data, nil
}

func (c *APIClient) CreateReply(replyInput CreateReplyInput) (Reply, error) {

	writeInCLI := replyInput.Content == "" //Check if the contents of the post have been handed in via argument. If not, use terminal text editor to write post.
	if writeInCLI {
		replyInput.Content = WriteContent() //See: utilities

	}
	if writeInCLI {
		if ConfirmPostIntention() == false {
			return Reply{}, nil
		}
	}

	replyJson, err := json.Marshal(replyInput)
	if err != nil {
		panic(err)
	}
	req, err := makeRequest("POST", c.ApiUrl+"/replies", c.Tokens, bytes.NewBuffer(replyJson))
	if err != nil {
		return Reply{}, fmt.Errorf("Error making reply request:%s", err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return Reply{}, fmt.Errorf("Error sending reply request:%s", err)
	}
	//fmt.Print(res.Status)
	//fmt.Print(res.Header)

	var replyConfirm createReplyResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&replyConfirm)
	if err != nil {
		return Reply{}, fmt.Errorf("Error decoding reply json:%s", err)
	}
	//fmt.Print(replyConfirm)

	reply := Reply{
		Content: replyInput.Content, PostID: replyInput.PostID, ParentReplyID: replyInput.ParentReplyID,
		CreatedAt: time.Now(), AuthorUsername: c.Username, ReplyID: replyConfirm.Data.ReplyID,
	}
	c.ReplyCache[reply.ReplyID] = reply

	return reply, nil
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
		//fmt.Printf("The reply was successfully deleted.\n")
	} else if res.StatusCode == 404 {
		return fmt.Errorf("No reply with that id found.\n")
	} else if res.StatusCode == 403 {
		return fmt.Errorf("You do not have authority to delete this reply.\n")
	} else {
		return fmt.Errorf("Something went wrong.\n")
	}
	return nil
}
