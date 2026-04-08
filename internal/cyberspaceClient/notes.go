package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GetNotesResponse struct {
	Data   []Note `json:"data"`
	Cursor string `json:"cursor"`
}

type OneNoteResponse struct {
	Data Note `json:"data"`
}

type Note struct {
	NoteID         string    `json:"noteId"`
	RevisionNumber int       `json:"revisionNumber"`
	AuthorID       string    `json:"authorId"`
	Content        string    `json:"content"`
	Deleted        bool      `json:"deleted"`
	Topics         []string  `json:"topics,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type CreateNoteInput struct {
	Content string   `json:"content"`
	Topics  []string `json:"topics,omitempty"`
}

func (c *APIClient) GetNotes(limit int, cursor string) (posts []Note, newCursor string, err error) {
	url := makeGetUrl(c.ApiUrl+"/notes", limit, cursor)

	req, err := makeRequest("GET", url, c.Tokens, nil)
	if err != nil {
		return []Note{}, cursor, fmt.Errorf("Error forming request: %s", err)
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return nil, cursor, fmt.Errorf("Error retrieving Notes: %s", err)
	}

	var getNotesResponse GetNotesResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&getNotesResponse)
	if err != nil {
		panic(err)
	}
	//fmt.Print(getNotificationsReply)
	for _, note := range getNotesResponse.Data {
		c.NoteCache[note.NoteID] = note
	}

	c.Cursors["posts_standard"] = getNotesResponse.Cursor
	return getNotesResponse.Data, getNotesResponse.Cursor, nil

}

func (c *APIClient) GetNoteById(note_id string) (Note, error) {
	req, err := makeRequest("GET", "https://api.cyberspace.online/v1/notes/"+note_id, c.Tokens, nil)
	if err != nil {
		return Note{}, fmt.Errorf("Error forming request: %s", err)
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return Note{}, fmt.Errorf("Error requesting post by ID: %s", err)
	}
	var oneNote OneNoteResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&oneNote)
	if err != nil {
		panic(err)
	}
	//fmt.Print(oneNote)
	return oneNote.Data, nil
}

func (c *APIClient) CreateNote(noteInput CreateNoteInput) (Note, error) {
	if noteInput.Content == "" {
		content := WriteContent()         //See: utilities
		topics := WriteTopics([]string{}) //See: utilities
		noteInput = CreateNoteInput{
			Content: content,
			Topics:  topics,
		}
	}

	postJson, err := json.Marshal(noteInput)
	if err != nil {
		panic(err)
	}
	req, err := makeRequest("POST", c.ApiUrl+"/notes", c.Tokens, bytes.NewBuffer(postJson))
	if err != nil {
		return Note{}, fmt.Errorf("Error making post request:%s", err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return Note{}, fmt.Errorf("Error sending post request:%s", err)
	}

	var noteConfirm struct {
		Data struct {
			NoteID string `json:"noteId"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&noteConfirm)
	if err != nil {
		return Note{}, fmt.Errorf("Error decoding post json:%s", err)
	}
	//fmt.Print(postConfirm)
	noteMade := Note{ //The response is just a post ID, so it's necessary to manually create the Post object for rendering on the client side. The alternative is to request the post from the server, but for optimization reasons (and becasue that is not possible for replies) I'm doing it the direct way.
		Content: noteInput.Content, Topics: noteInput.Topics, NoteID: noteConfirm.Data.NoteID,
		//IsPublic: postInput.IsPublic, IsNSFW: postInput.IsNSFW,
	}
	c.NoteCache[noteMade.NoteID] = noteMade
	return noteMade, nil
}

func (c *APIClient) UpdateNote(noteInput CreateNoteInput, noteID string) (Note, error) {
	if noteInput.Content == "" {
		content := WriteContent()         //See: utilities
		topics := WriteTopics([]string{}) //See: utilities
		noteInput = CreateNoteInput{
			Content: content,
			Topics:  topics,
		}
	}

	postJson, err := json.Marshal(noteInput)
	if err != nil {
		panic(err)
	}
	req, err := makeRequest("PATCH", c.ApiUrl+"/notes/"+noteID, c.Tokens, bytes.NewBuffer(postJson))
	if err != nil {
		return Note{}, fmt.Errorf("Error making post request:%s", err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return Note{}, fmt.Errorf("Error sending post request:%s", err)
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != 201 {
		fmt.Print(res.StatusCode)
		var errResp ErrorResponse
		decoder := json.NewDecoder(res.Body)
		err = decoder.Decode(&errResp)

		fmt.Print(errResp)
	}

	var noteConfirm struct {
		Data struct {
			NoteID string `json:"noteId"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&noteConfirm)
	if err != nil {
		return Note{}, fmt.Errorf("Error decoding post json:%s", err)
	}
	//fmt.Print(postConfirm)
	noteUpdated := Note{ //The response is just a post ID, so it's necessary to manually create the Post object for rendering on the client side. The alternative is to request the post from the server, but for optimization reasons (and becasue that is not possible for replies) I'm doing it the direct way.
		Content: noteInput.Content, Topics: noteInput.Topics, NoteID: noteConfirm.Data.NoteID,
		//IsPublic: postInput.IsPublic, IsNSFW: postInput.IsNSFW,
	}
	return noteUpdated, nil
}

func (c *APIClient) DeleteNote(noteId string) error {

	req, err := makeRequest("DELETE", c.ApiUrl+"/notes/"+noteId, c.Tokens, nil)
	if err != nil {
		return fmt.Errorf("Error forming delete request: %s", err)
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return fmt.Errorf("Error during the request process: %s", err)
	}
	if res.StatusCode == 200 || res.StatusCode == 201 {
		return nil
	} else {
		return fmt.Errorf("Something went wrong: %s", res.Status)
	}
}
