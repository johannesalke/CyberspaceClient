package client

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
