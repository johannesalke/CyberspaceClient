package main

import (
	"bufio"
	//"bytes"
	//"encoding/json"
	"fmt"
	"net/http"
	"os"
	//"os/exec"
	"strings"
	//"time"

	client "github.com/johannesalke/CyberspaceClient/internal/cyberspaceClient"
)

type Config struct {
	apiUrl   string
	cache    map[string]any
	tokens   client.AuthTokens
	username string
	client   http.Client
}

func main() {

	csc := client.InitAPIClient()

	//cfg := Config{apiUrl: "https://api.cyberspace.online/v1"}
	//client := http.NewClientHandler()
	csc.Tokens = client.Login(csc.ApiUrl)
	fmt.Printf("authToken: %.10s", csc.Tokens.IDToken)
	posts, _, err := csc.GetPosts(5, "")
	if err != nil {
		fmt.Print("Error: ", err)
	}
	for _, post := range posts {
		if post.IsNSFW == true {
			continue
		}
		fmt.Print(post.AuthorUsername, " | ", post.PostID, "\n")
		fmt.Print("=======================================================", "\n")
		fmt.Print(post.Content, "\n")
		fmt.Print("=======================================================", "\n")
	}
	//client.Post{}
	id := "Leg8tjQYjTZo9cOySqb4"
	post, err := csc.GetPostById(id)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(post.AuthorUsername, post.Content)
	err = csc.DeletePost(id)
	if err != nil {
		fmt.Print(err)
	}
	for true {
		x := 5
		x = x + 5
	}
	scanner := bufio.NewScanner(os.Stdin)
	err = csc.CreatePost(csc.Tokens)
	if err != nil {
		fmt.Print(err)
	}
	for true {
		scanner.Scan()
		input := scanner.Text()
		args := strings.Split(input, " ")
		if len(args) == 0 {
			continue
		}

		//cmd := args[0]

	}

}

////////////////////////////| Posts |///////////////////////////

/*
func (cfg *Config) sendRequest() {
	body := []byte(`{"name":"John"}`)

	req, err := makeRequest()
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer ") //+cfg.tokens.IDToken

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func

	res, err := http.DefaultClient.Do(req)

	if res.StatusCode == 401 {
		tokens = auth.TokenRefresh(url, tokens)
		req, err := http.NewRequest(method, url, body)
		req.Header.Set("Authorization", "Bearer "+tokens.IDToken)

		res, err = http.DefaultClient.Do(req)

		if res.StatusCode
	}
	return req, nil
*/
