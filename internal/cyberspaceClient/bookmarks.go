package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type getBooksmarksResponse struct {
	Data   []Bookmark `json:"data"`
	Cursor string     `json:"cursor"`
}

type Bookmark struct {
	BookmarkID string    `json:"bookmarkId"`
	UserID     string    `json:"userId"`
	PostID     string    `json:"postId,omitempty"`
	Type       string    `json:"type"`
	CreatedAt  time.Time `json:"createdAt"`
	ReplyID    string    `json:"replyId,omitempty"`
}

func (c *APIClient) GetBookmarks(limit int, cursor string) (posts []Bookmark, newCursor string, err error) {
	url := makeGetUrl(c.ApiUrl+"/bookmarks", limit, cursor)

	req, err := makeRequest("GET", url, c.Tokens, nil)
	if err != nil {
		return []Bookmark{}, cursor, fmt.Errorf("Error forming request: %s", err)
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return nil, cursor, fmt.Errorf("Error retrieving Bookmarks: %s", err)
	}
	if err := c.expectSuccess(res, "retrieving bookmarks"); err != nil {
		return nil, cursor, err
	}

	var getBooksmarksResponse getBooksmarksResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&getBooksmarksResponse)
	if err != nil {
		return nil, cursor, fmt.Errorf("error decoding bookmarks: %s", err)
	}
	for _, bookmark := range getBooksmarksResponse.Data {
		c.BookmarkCache[bookmark.BookmarkID] = bookmark
	}

	//fmt.Print(getNotificationsReply)
	c.Cursors["bookmarks"] = getBooksmarksResponse.Cursor
	return getBooksmarksResponse.Data, getBooksmarksResponse.Cursor, nil

}

type CreateBookmarkInput struct {
	ReplyID string `json:"replyId,omitempty"`
	PostID  string `json:"postId,omitempty"`
	Type    string `json:"type"`
}

type createBookmarkConfirm struct {
	Data struct {
		BookmarkID string `json:"bookmarkId"`
	} `json:"data"`
}

func (c *APIClient) CreateBookmark(id, bookmarkType string) error {
	bookmarkInput := CreateBookmarkInput{Type: bookmarkType}
	if bookmarkType == "post" {
		bookmarkInput.PostID = id
	} else if bookmarkType == "reply" {
		bookmarkInput.ReplyID = id
	} else {
		return fmt.Errorf("Invalid type of bookmarked object. Must be either 'post' or 'reply'")
	}

	postJson, err := json.Marshal(bookmarkInput)
	if err != nil {
		return fmt.Errorf("error encoding bookmark: %s", err)
	}
	req, err := makeRequest("POST", c.ApiUrl+"/bookmarks", c.Tokens, bytes.NewBuffer(postJson))
	if err != nil {
		return fmt.Errorf("Error making post request:%s", err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("Error sending post request:%s", err)
	}
	return c.expectSuccess(res, "creating bookmark")
	/*
		var bookmarkConfirm createBookmarkConfirm
		decoder := json.NewDecoder(res.Body)
		err = decoder.Decode(&bookmarkConfirm)
		if err != nil {
			return fmt.Errorf("Error decoding bookmark json:%s", err)
		}*/
	//fmt.Print(postConfirm)
	//fmt.Print(res.Status)
	//fmt.Print(res.Header)

}

func (c *APIClient) DeleteBookmark(bookmarkId string) error {

	req, err := makeRequest("DELETE", c.ApiUrl+"/bookmarks/"+bookmarkId, c.Tokens, nil)
	if err != nil {
		return fmt.Errorf("Error forming delete request: %s", err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("Error during the request process: %s", err)
	}
	return c.expectSuccess(res, "deleting bookmark")
}
