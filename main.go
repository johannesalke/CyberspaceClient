package main

import (
	"bufio"
	//"github.com/fatih/color"
	"os/signal"
	"syscall"
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
	//color.Set(color.BgHiGreen)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		fmt.Print("\033[0m") // Reset on interrupt
		fmt.Print("\n")
		os.Exit(0)
	}()
	fmt.Print(`
	 ██████╗██╗   ██╗██████╗ ███████╗██████╗ ███████╗██████╗  █████╗  ██████╗███████╗
	██╔════╝╚██╗ ██╔╝██╔══██╗██╔════╝██╔══██╗██╔════╝██╔══██╗██╔══██╗██╔════╝██╔════╝
	██║      ╚████╔╝ ██████╔╝█████╗  ██████╔╝███████╗██████╔╝███████║██║     █████╗
	██║       ╚██╔╝  ██╔══██╗██╔══╝  ██╔══██╗╚════██║██╔═══╝ ██╔══██║██║     ██╔══╝
	╚██████╗   ██║   ██████╔╝███████╗██║  ██║███████║██║     ██║  ██║╚██████╗███████╗
	 ╚═════╝   ╚═╝   ╚═════╝ ╚══════╝╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝  ╚═╝ ╚═════╝╚══════╝
`)

	defer fmt.Print("\033[0m")
	//fmt.Print("\172[0m") fmt.Print("\033[38;5;203m")
	fmt.Print("\033[38;5;223m")

	var csc = client.InitAPIClient()
	//fmt.Print(csc)
	csc.Config = client.GetConfig()
	//fmt.Print(csc.Config)

	//cfg := Config{apiUrl: "https://api.cyberspace.online/v1"}
	//client := http.NewClientHandler()
	if csc.Config.StayLoggedIn == true {
		csc.Tokens = client.AuthTokens{RefreshToken: "", IDToken: "", RTDBToken: ""}
		csc.Tokens.RefreshToken = csc.Config.StoredValues.RefreshToken
		//fmt.Print((csc.Tokens.RefreshToken), "\n")
		csc.TokenRefresh()
		fmt.Print("You are still logged in.\n")

	} else {
		csc.Tokens = client.Login(csc.ApiUrl)
	}

	fmt.Printf("[authToken: %.10s...]\n", csc.Tokens.IDToken)

	c := commands{make(map[string]func(*client.APIClient, command) error)}
	c.register("view", handlerView)
	c.register("write", handlerWrite)
	c.register("note", handlerUpdateNote)
	//c.register("config", handlerUpdateConfig)

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
		fmt.Print("\033[38;5;223m")
		if err != nil {
			fmt.Println(err)
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

//=====================|Level 1 Handlers|=========================

func handlerView(csc *client.APIClient, cmd command) error { // Redirects to handlers: viewFeed, viewPost, viewNotes, view Notifications, ...
	if len(cmd.Args) == 0 {
		renderPrint("The 'view' command requires an argument. Valid arguments: feed, post <id>, notifications, notes.\n")
		return nil
	}

	switch cmd.Args[0] {
	case "feed":
		return handlerViewFeed(csc, cmd)
	case "notifications":
		return handlerViewNotifications(csc, cmd)
	case "post":
		return handlerViewPost(csc, cmd)
	case "notes":
		return handlerViewNotes(csc, cmd)

	}
	return nil
}

func handlerWrite(csc *client.APIClient, cmd command) error {
	if len(cmd.Args) == 0 {
		renderPrint("The 'write' command requires an argument. Valid arguments: post, note\n")
		return nil
	}

	switch cmd.Args[0] {
	case "post":
		return handlerWritePost(csc, cmd)

	case "note":
		return handlerViewNotifications(csc, cmd)
	case "":
		return handlerViewPost(csc, cmd)
	case "notes":
		return handlerViewNotes(csc, cmd)

	}
	return nil
}

//////////////////| View Handlers |///////////////////////////////

func handlerViewFeed(csc *client.APIClient, cmd command) error {

	posts, _, err := csc.GetPosts(5, csc.Cursors["feed"])
	if err != nil {
		return err
	}
	for _, post := range posts {
		if post.IsNSFW == true {
			continue
		}
		renderPost(post, false)

	}
	return nil
}

func handlerViewPost(csc *client.APIClient, cmd command) error {

	post_id := cmd.Args[0]
	post, err := csc.GetPostById(post_id)
	if err != nil {
		fmt.Print(err)
	}
	renderPost(post, true)
	replies, _, err := csc.GetReplies(post_id, 20, "")
	if err != nil {
		fmt.Print(err)
	}

	for _, reply := range replies {

		renderReply(reply)

	}

	if err != nil {
		fmt.Print(err)
	}
	return nil
}

func handlerViewNotifications(csc *client.APIClient, cmd command) error {

	notifications, new_cursor, err := csc.GetNotifications(10, csc.Cursors["notifications"])
	if err != nil {
		fmt.Printf("Error getting notifs: %s", err)
	}
	csc.Cursors["notifications"] = new_cursor
	for _, notification := range notifications {
		renderNotification(csc, notification)
	}
	return nil
}

func handlerViewNotes(csc *client.APIClient, cmd command) error {
	return nil
}

///////////////| Writing Handlers |////////////////////////

func handlerWritePost(csc *client.APIClient, cmd command) error {
	err := csc.CreatePost()
	if err != nil {
		fmt.Print(err)
	}
	return nil
}

////////////////| Editing Handlers |////////////////////////////

func handlerUpdateConfig(csc *client.APIClient, cmd command) error {

	csc.UpdateConfig()
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
