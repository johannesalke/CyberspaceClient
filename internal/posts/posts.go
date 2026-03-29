package posts

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/johannesalke/CyberspaceTUI/internal/auth"
)

func GetPosts(tokens auth.AuthTokens, limit int, from_id string) {
	var body bytes.Buffer
	http.NewRequest("GET", "https://api.cyberspace.online/v1/posts", body)
}

func CreatePost(tokens auth.AuthTokens, message string) {

}

func DeletePost(tokens auth.AuthTokens, postId string) error {

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
