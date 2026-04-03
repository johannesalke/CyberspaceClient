package client

import (
	"encoding/json"
	"fmt"
	"time"
)

//State: Incomplete

type GetUserResponse struct {
	Data User `json:"data"`
}

type User struct {
	UserID            string    `json:"userId"`
	Username          string    `json:"username"`
	IsBanned          bool      `json:"isBanned,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	IsSupporter       bool      `json:"isSupporter"`
	PermissionImage   bool      `json:"permissionImage"`
	ProfilePictureURL string    `json:"profilePictureUrl"`
	UpdatedAt         time.Time `json:"updatedAt"`
	SupporterIcon     string    `json:"supporterIcon"`
	SerialNumber      int       `json:"serialNumber"`
	IsWikiEditor      bool      `json:"isWikiEditor,omitempty"`
	GuildSlug         string    `json:"guildSlug"`
	GuildID           string    `json:"guildId"`
	GuildIcon         string    `json:"guildIcon"`
	HasPublicPosts    bool      `json:"hasPublicPosts"`
	PinnedPostID      string    `json:"pinnedPostId"`
	Bio               string    `json:"bio"`
	FollowingCount    int       `json:"followingCount"`
	LastActiveAt      time.Time `json:"lastActiveAt"`
	FollowersCount    int       `json:"followersCount"`
	PostsCount        int       `json:"postsCount"`
	PublicPostsCount  int       `json:"publicPostsCount"`
	IsHacker          bool      `json:"isHacker,omitempty"`
	IsImmortal        bool      `json:"isImmortal,omitempty"`
	LocationLatitude  float64   `json:"locationLatitude,omitempty"`
	LocationLongitude float64   `json:"locationLongitude,omitempty"`
	WebsiteName       string    `json:"websiteName,omitempty"`
	WebsiteURL        string    `json:"websiteUrl,omitempty"`
	WebsiteImageURL   string    `json:"websiteImageUrl,omitempty"`
	LocationName      string    `json:"locationName,omitempty"`
}

func (c *APIClient) GetMyUserProfile() (User, error) {

	req, err := makeRequest("GET", "https://api.cyberspace.online/v1/users/me", c.Tokens, nil)
	if err != nil {
		return User{}, fmt.Errorf("Error forming request: %s", err)
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return User{}, fmt.Errorf("Error requesting post by ID: %s", err)
	}
	var userResponse GetUserResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&userResponse)
	if err != nil {
		panic(err)
	}
	//fmt.Print(userResponse.Data)
	return userResponse.Data, nil
}

func (c *APIClient) GetUserProfileByName(username string) (User, error) {

	req, err := makeRequest("GET", "https://api.cyberspace.online/v1/users/"+username, c.Tokens, nil)
	if err != nil {
		return User{}, fmt.Errorf("Error forming request: %s", err)
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return User{}, fmt.Errorf("Error requesting post by ID: %s", err)
	}
	var userResponse GetUserResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&userResponse)
	if err != nil {
		panic(err)
	}
	//fmt.Print(userResponse.Data)
	return userResponse.Data, nil
}

func (c *APIClient) GetUsersPosts(username string, limit int, cursor string) (posts []Post, newCursor string, err error) {
	url := makeGetUrl(c.ApiUrl+"users/"+username+"/posts", limit, cursor)

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
	cursor_key := "userposts_" + username
	c.Cursors[cursor_key] = getPostsResponse.Cursor
	return getPostsResponse.Data, getPostsResponse.Cursor, nil

}

func (c *APIClient) GetUserReplies(username string, limit int, cursor string) (replies []Reply, newCursor string, err error) {
	url := makeGetUrl(c.ApiUrl+"/users/"+username+"/replies", limit, cursor)

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
	cursor_key := "userreplies_" + username
	c.Cursors[cursor_key] = getRepliesResponse.Cursor

	return getRepliesResponse.Data, getRepliesResponse.Cursor, nil

}

//////////////| Profile updates left out for now|/////////////////////

/*
type UpdateProfileInput struct {
	Bio               string  `json:"bio"`               //null
	PinnedPostID      string  `json:"pinnedPostId"`      //null
	DisplayName       string  `json:"displayName"`       //null
	WebsiteURL        string  `json:"websiteUrl"`        //must start with http(s)://, null to clear
	WebsiteName       string  `json:"websiteName"`       //null to clear
	WebsiteImageURL   string  `json:"websiteImageUrl"`   //null to clear
	LocationLatitude  float64 `json:"locationLatitude"`  //null to clear
	LocationLongitude float64 `json:"locationLongitude"` //null to clear
	LocationName      string  `json:"locationName"`      //null to clear
}

type UpdateProfileRequest struct {
	Bio               *string  `json:"bio,omitempty"`
	PinnedPostID      *string  `json:"pinnedPostId,omitempty"`
	DisplayName       *string  `json:"displayName,omitempty"`
	WebsiteURL        *string  `json:"websiteUrl,omitempty"`
	WebsiteName       *string  `json:"websiteName,omitempty"`
	WebsiteImageURL   *string  `json:"websiteImageUrl,omitempty"`
	LocationLatitude  *float64 `json:"locationLatitude,omitempty"`
	LocationLongitude *float64 `json:"locationLongitude,omitempty"`
	LocationName      *string  `json:"locationName,omitempty"`
}

func (c *APIClient) UpdateOwnProfile()
*/
