package main

import (
	"bufio"
	"maps"
	"slices"
	"strconv"

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

	client "github.com/johannesalke/cyberspacecli/internal/cyberspaceClient"
)

type Config struct {
	apiUrl   string
	cache    map[string]any
	tokens   client.AuthTokens
	username string
	client   http.Client
}

var IDmap = make(map[int]string)
var reverseIDmap = make(map[string]int)

//

func main() {
	//fmt.Print(err)
	IDmap[0] = "existence"
	reverseIDmap["nonexistence"] = 0

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
	fmt.Print("\033[38;5;172m")

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
	c.register("edit", handlerEdit)
	c.register("publish", handlerPublish)
	//c.register("config", handlerUpdateConfig)

	scanner := bufio.NewScanner(os.Stdin)

	for true {

		fmt.Print("> ")
		scanner.Scan()
		input := scanner.Text()
		arguments := strings.Split(input, " ")
		if len(arguments) == 0 {
			continue
		} else if arguments[0] == "exit" {
			break
		}
		cmd := command{Name: arguments[0], Args: arguments[1:]}
		err := c.run(&csc, cmd)
		if csc.LastStatusCode == 401 {
			csc.TokenRefresh()
			err = c.run(&csc, cmd)
		}
		fmt.Print("\033[38;5;172m")
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
	default:
		return fmt.Errorf("Unknown argument. Valid arguments for view: feed, post <id>, notifications, notes.\n")
	}

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
		return handlerWriteNote(csc, cmd)
	default:
		return fmt.Errorf("Unknown argument. Valid arguments for write: post, note.\n")

	}

}

func handlerEdit(csc *client.APIClient, cmd command) error {
	if len(cmd.Args) == 0 {
		renderPrint("The 'edit' command requires an argument. Valid arguments: note <note_id>, config\n")
		return nil
	}

	switch cmd.Args[0] {
	case "note":
		return handlerEditNote(csc, cmd)
	case "config":
		return handlerEditConfig(csc, cmd)
	default:
		return fmt.Errorf("Unknown argument. Valid arguments for write: post, note.\n")

	}

}

func handlerPublish(csc *client.APIClient, cmd command) error {
	if len(cmd.Args) != 2 {
		renderPrint("The 'publish' command requires two extra arguments: note & <note_id>\n")
		return nil
	}

	switch cmd.Args[0] {
	case "note":
		return handlerPublishNote(csc, cmd)
	default:
		return fmt.Errorf("Unknown argument. Valid arguments for publish: note.\n")

	}

}

//////////////////| View Handlers |///////////////////////////////

func handlerViewFeed(csc *client.APIClient, cmd command) error {
	if len(cmd.Args) == 2 && cmd.Args[1] == "new" { //Check for new posts rather than going further down the feed
		cursor_temp := csc.Cursors["feed"]
		var old_posts = false
		for i := 0; !old_posts && i < 5; i++ { //Gets up to 15 posts from the start of the feed. If any set of 3 includes old posts, stop getting posts.
			posts, _, err := csc.GetPosts(4, "")
			if err != nil {
				return err
			}
			for _, post := range posts {
				if post.IsNSFW == true {
					continue
				}
				renderPost(post, false)
				_, old_posts = simplifyID(post.PostID) //Checks if this iteration crossed into new posts.
			}
		}
		csc.Cursors["feed"] = cursor_temp
		return nil
	} else if len(cmd.Args) == 2 && cmd.Args[1] == "reset" { //Permanently reset the cursor.
		csc.Cursors["feed"] = ""
	}

	posts, _, err := csc.GetPosts(10, csc.Cursors["feed"]) //Normal feed viewing.
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
} // Complete ~

func handlerViewPost(csc *client.APIClient, cmd command) error {

	post_id := cmd.Args[1]

	fullPostID, err := getFullID(post_id)
	if err != nil {
		fmt.Print(err)
	}
	post, err := csc.GetPostById(fullPostID)
	if err != nil {
		fmt.Print(err)
	}
	renderPost(post, true)
	replies, _, err := csc.GetReplies(fullPostID, 20, "")
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
} // Complete ~

func handlerViewNotifications(csc *client.APIClient, cmd command) error {
	if len(cmd.Args) == 2 && cmd.Args[1] == "new" { //Check for new notifs rather than going further down the feed
		cursor_temp := csc.Cursors["notifications"]
		notifications, _, err := csc.GetNotifications(10, "")
		if err != nil {
			fmt.Printf("Error getting notifs: %s", err)
		}
		for _, notification := range notifications {
			renderNotification(csc, notification)
		}
		csc.Cursors["notifications"] = cursor_temp
		return nil
	} else if len(cmd.Args) == 2 && cmd.Args[1] == "reset" { //reset the notification cursor.
		csc.Cursors["notifications"] = ""
	}

	notifications, new_cursor, err := csc.GetNotifications(10, csc.Cursors["notifications"])
	if err != nil {
		fmt.Printf("Error getting notifs: %s", err)
	}
	csc.Cursors["notifications"] = new_cursor
	for _, notification := range notifications {
		renderNotification(csc, notification)
	}
	return nil
} // Complete ~

func handlerViewNotes(csc *client.APIClient, cmd command) error {
	notes, _, err := csc.GetNotes(10, csc.Cursors["notes"])
	if err != nil {
		return err
	}
	var already_displayed_notes []string //If a note was edited before, the List Notes API will send you all versions of it. This counter is there to ensure only the most up-to-date version is displayed.
	for _, note := range notes {
		if slices.Contains(already_displayed_notes, note.NoteID) {
			continue //Skip notes that already had a newer version displayed
		}
		renderNote(note, true)
		already_displayed_notes = append(already_displayed_notes, note.NoteID)

	}
	return nil
} // Complete ~

func handlerViewBookmarks(csc *client.APIClient, cmd command) error {

	return nil
} // Empty

///////////////| Writing Handlers |////////////////////////

func handlerWritePost(csc *client.APIClient, cmd command) error {
	post, err := csc.CreatePost(client.CreatePostInput{})
	if err != nil {
		fmt.Print(err)
	}
	if post.Content == "" { //If the user either wrote nothing in the document, or didn't confirm intention to post.
		return nil
	}
	renderPost(post, true)
	return nil
} //|Complete

func handlerWriteNote(csc *client.APIClient, cmd command) error {
	note, err := csc.CreateNote(client.CreateNoteInput{})
	if err != nil {
		fmt.Print(err)
	}
	renderNote(note, true)
	return nil
} //|Complete

////////////////| Editing Handlers |////////////////////////////

func handlerEditConfig(csc *client.APIClient, cmd command) error {

	csc.UpdateConfig()
	return nil
} //|Complete

func handlerEditNote(csc *client.APIClient, cmd command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("This command requiers an additional argument: The id of the note to be edited.")
	}
	note_id := cmd.Args[1]

	fullNoteID, err := getFullID(note_id)
	Note, err := csc.GetNoteById(fullNoteID)
	if err != nil {
		return fmt.Errorf("Error: %s ", err)

	}
	newNoteInput, err := client.EditNote(Note)
	if err != nil {
		return fmt.Errorf("Error: %s ", err)

	}
	newNote, err := csc.UpdateNote(newNoteInput, fullNoteID)
	if err != nil {
		return fmt.Errorf("Error: %s ", err)

	}

	renderNote(newNote, true)
	fmt.Print(newNote.NoteID, "\n")
	return nil
} //|Complete

////////////////| Publish Handler |////////////////////////////

func handlerPublishNote(csc *client.APIClient, cmd command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("This command requiers an additional argument: The id of the note to be published.")
	}
	note_id := cmd.Args[1]

	fullNoteID, err := getFullID(note_id)
	note, err := csc.GetNoteById(fullNoteID)
	if err != nil {
		return fmt.Errorf("Error: %s ", err)

	}
	postInput := client.CreatePostInput{Content: note.Content, Topics: note.Topics}
	post, err := csc.CreatePost(postInput)
	if err != nil {
		return fmt.Errorf("Error publishing note: %s", err)
	}
	renderPost(post, true)
	return nil
}

////////////////////| id utilities |////////////////////////////////

//var IDmap = make(map[int]string)
//var reverseIDmap = make(map[string]int)

func simplifyID(fullID string) (simpleID int, exists bool) {
	currentValue := reverseIDmap[fullID] //Check if post already exists in database
	if currentValue != 0 {
		//fmt.Print("Id already exists, fam.")
		return currentValue, true
	}

	//If it does not already exists
	idKeys := maps.Keys(IDmap)
	var idKeysSlice []int
	for key := range idKeys {
		idKeysSlice = append(idKeysSlice, key)
	}

	maxValue := slices.Max(idKeysSlice)
	newSimpleID := maxValue + 1
	IDmap[newSimpleID] = fullID
	reverseIDmap[fullID] = newSimpleID
	return newSimpleID, false
}

func getFullID(simpleID string) (fullID string, err error) {
	simpleIDString, err := strconv.Atoi(simpleID)
	if err != nil {
		fmt.Print(err)
	}
	fullID = IDmap[simpleIDString]
	if fullID == "" {
		return "", fmt.Errorf("There is no object with this id")
	}
	return fullID, nil

}
