package client

import (
	"encoding/json"
	"fmt"
	"time"
)

type NotificationType int

const (
	Replied    NotificationType = iota // 0
	Followed                           // 1
	Bookmarked                         // 2
	Posted                             // 3
	Poked
)

type GetNotificationsReply struct {
	Data   []Notification `json:"data"`
	Cursor string         `json:"cursor"`
}

type Notification struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	ActorID       string `json:"actorId"`
	ActorUsername string `json:"actorUsername"`
	TargetID      string `json:"targetId"`   //Who is being replied to?
	TargetType    string `json:"targetType"` //Reply, Follow, etc.
	Metadata      struct {
		AuthorUsername string `json:"authorUsername"`
		ReplyID        string `json:"replyId"`
	} `json:"metadata"`
	UserID    string    `json:"userId"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"createdAt"`
}

type MarkAllReadReply struct {
	Data struct {
		Updated int `json:"updated"`
	} `json:"data"`
}

func (c *APIClient) GetNotifications(limit int, cursor string) (notifications []Notification, newCursor string, err error) {
	url := makeGetUrl(c.ApiUrl+"/notifications", limit, cursor)

	req, err := makeRequest("GET", url, c.Tokens, nil)
	if err != nil {
		return nil, cursor, fmt.Errorf("Error forming request: %s", err)
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return nil, cursor, fmt.Errorf("Error retrieving Notifications: %s", err)
	}
	if err := c.expectSuccess(res, "retrieving notifications"); err != nil {
		return nil, cursor, err
	}

	var getNotificationsReply GetNotificationsReply
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&getNotificationsReply)
	if err != nil {
		return nil, cursor, fmt.Errorf("error decoding notifications: %s", err)
	}
	//fmt.Print(getNotificationsReply)
	c.Cursors["notifications_standard"] = getNotificationsReply.Cursor
	return getNotificationsReply.Data, getNotificationsReply.Cursor, nil
}

func (c *APIClient) MarkAsRead(notificationID string) error {
	req, err := makeRequest("PATCH", c.ApiUrl+"/notifications/"+notificationID, c.Tokens, nil)
	if err != nil {
		return fmt.Errorf("Error forming MarkAsRead request: %s", err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return err
	}
	return c.expectSuccess(res, "marking notification read")
}

func (c *APIClient) MarkAllAsRead() (int, error) {
	req, err := makeRequest("PATCH", c.ApiUrl+"/notifications/read-all", c.Tokens, nil)
	if err != nil {
		return 0, fmt.Errorf("Error forming MarkAllAsRead request: %s", err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return 0, err
	}
	if err := c.expectSuccess(res, "marking notifications read"); err != nil {
		return 0, err
	}
	var markAllReadReply MarkAllReadReply
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&markAllReadReply)
	if err != nil {
		return 0, fmt.Errorf("error decoding read status: %s", err)
	}

	return markAllReadReply.Data.Updated, nil
}
