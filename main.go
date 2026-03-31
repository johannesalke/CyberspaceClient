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

	//glamour "charm.land/glamour/v2"

	client "github.com/johannesalke/CyberspaceClient/internal/cyberspaceClient"
)

type Config struct {
	apiUrl   string
	cache    map[string]any
	tokens   client.AuthTokens
	username string
	client   http.Client
}

//

func main() {
	//fmt.Print(err)

	//renderer, _ := glamour.NewTermRenderer(glamour.WithStylePath("dark"))
	//out, _ := renderer.Render("# Heading\n\n**Bold text**\n\n- List item")
	//fmt.Print(out)

	var csc = client.InitAPIClient()
	//fmt.Print(csc)
	//csc.Config = client.GetConfig()
	//fmt.Print(csc.Config)

	//cfg := Config{apiUrl: "https://api.cyberspace.online/v1"}
	//client := http.NewClientHandler()
	/*if csc.Config.StayLoggedIn == true {
		csc.Tokens = client.AuthTokens{RefreshToken: "", IDToken: "", RTDBToken: ""}
		csc.Tokens.RefreshToken = csc.Config.StoredValues.RefreshToken
		fmt.Print((csc.Tokens.RefreshToken), "\n")
		csc.TokenRefresh()
	} else {

	}*/
	csc.Tokens = client.Login(csc.ApiUrl)
	fmt.Printf("authToken: %.10s |\n", csc.Tokens.IDToken)

	/*id := "nxSSfugK6L9tFBSF1zEZ"

	fmt.Print(id)
	os.Exit(0)*/

	//client.Post{}
	c := commands{make(map[string]func(*client.APIClient, command) error)}
	c.register("feed", handlerViewFeed)
	c.register("write", handlerCreatePost)
	c.register("replies", handlerViewPost)
	c.register("note", handlerUpdateNote)
	//c.register("config", handlerUpdateConfig)
	/*
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

		err = csc.CreatePost()
		if err != nil {
			fmt.Print(err)
		}
	*/
	scanner := bufio.NewScanner(os.Stdin)

	for true {

		fmt.Print("> ")
		scanner.Scan()
		input := scanner.Text()
		arguments := strings.Split(input, " ")
		if len(arguments) == 0 {
			continue
		}
		cmd := command{Name: arguments[0], Args: arguments[1:]}
		err := c.run(&csc, cmd)
		if csc.LastStatusCode == 401 {
			csc.TokenRefresh()
			err = c.run(&csc, cmd)
		}
		if err != nil {
			fmt.Println(&err)
		}

		//cmd := args[0]

	}

}

//==========================================================================================

type command struct {
	Name string
	Args []string
}

type commands struct {
	commands map[string]func(*client.APIClient, command) error
}

func (c *commands) run(s *client.APIClient, cmd command) error {
	if cmdFunc, ok := c.commands[cmd.Name]; ok {
		return cmdFunc(s, cmd)
	}
	return fmt.Errorf("Error: Command used not registered. ")
}
func (c *commands) register(name string, f func(*client.APIClient, command) error) {
	c.commands[name] = f
}

///=======================================

func handlerViewFeed(csc *client.APIClient, cmd command) error {

	posts, _, err := csc.GetPosts(5, csc.Cursors["feed"])
	if err != nil {
		return err
	}
	for _, post := range posts {
		if post.IsNSFW == true {
			continue
		}
		renderPost(post)

	}
	return nil
}

func handlerViewPost(csc *client.APIClient, cmd command) error {

	post_id := cmd.Args[0]
	post, err := csc.GetPostById(post_id)
	if err != nil {
		fmt.Print(err)
	}
	renderPost(post)
	replies, _, err := csc.GetReplies(post_id, 20, "")
	if err != nil {
		fmt.Print(err)
	}

	for _, reply := range replies {

		renderReplies(reply)

	}

	if err != nil {
		fmt.Print(err)
	}
	return nil
}

func handlerCreatePost(csc *client.APIClient, cmd command) error {
	err := csc.CreatePost()
	if err != nil {
		fmt.Print(err)
	}
	return nil
}

func handlerUpdateConfig(csc *client.APIClient, cmd command) error {

	csc.UpdateConfig()
	return nil
}

func handlerViewNotifications(csc *client.APIClient, cmd command) error {

	return nil
}

func handlerUpdateNote(csc *client.APIClient, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("This command requiers one argument: The idea of the note to be updated.")
	}

	Note, err := csc.GetNoteById(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Error: %s ", err)

	}
	newNote, err := client.EditNote(Note)
	if err != nil {
		return fmt.Errorf("Error: %s ", err)

	}
	id, err := csc.UpdateNote(newNote, cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Error: %s ", err)

	}
	fmt.Print(id, "\n")
	return nil
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
