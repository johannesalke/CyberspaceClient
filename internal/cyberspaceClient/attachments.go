package client

import "encoding/json"

type ImgAttachment struct {
	Type   string `json:"type"`
	Src    string `json:"src"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type AudioAttachment struct {
	Type   string `json:"type"`
	Src    string `json:"src"`
	Origin string `json:"origin"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
	Genre  string `json:"genre"`
}

type Attachment struct {
	Type   string `json:"type"`
	Src    string `json:"src"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Origin string `json:"origin,omitempty"`
	Artist string `json:"artist,omitempty"`
	Title  string `json:"title,omitempty"`
	Genre  string `json:"genre,omitempty"`
}

// AudioAttachments returns the playable audio attachments carried by a post.
// Post.Attachments arrives as untyped JSON, so it is re-encoded and decoded
// into a concrete slice before filtering.
func (p Post) AudioAttachments() ([]AudioAttachment, error) {
	if p.Attachments == nil {
		return nil, nil
	}
	raw, err := json.Marshal(p.Attachments)
	if err != nil {
		return nil, err
	}
	var all []Attachment
	if err := json.Unmarshal(raw, &all); err != nil {
		return nil, err
	}
	var audio []AudioAttachment
	for _, a := range all {
		if a.Type != "audio" {
			continue
		}
		audio = append(audio, AudioAttachment{
			Type:   a.Type,
			Src:    a.Src,
			Origin: a.Origin,
			Artist: a.Artist,
			Title:  a.Title,
			Genre:  a.Genre,
		})
	}
	return audio, nil
}
