package main

import (
	"bufio"
	"maps"
	"slices"
	"strconv"
	"time"

	//"github.com/fatih/color"
	"os/signal"
	"syscall"

	//"bytes"
	//"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	//"os/exec"
	"strings"
	//"time"
	//"golang.org/x/sys/windows"

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

func main() {

	operatingSystem := runtime.GOOS

	//fmt.Print(err)
	IDmap[0] = "existence"
	reverseIDmap["nonexistence"] = 0

	//renderer, _ := glamour.NewTermRenderer(glamour.WithStylePath("dark"))
	//out, _ := renderer.Render("# Heading\n\n**Bold text**\n\n- List item")
	//fmt.Print(out)
	//color.Set(color.BgHiGreen)
	//EnableANSI()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		if operatingSystem != "windows" {
		fmt.Print("\033[0m") // Reset on interrupt
		}
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

	time.Sleep(500 * time.Millisecond)
	defer fmt.Print("\033[0m")
	//fmt.Print("\172[0m") fmt.Print("\033[38;5;203m")
	if operatingSystem != "windows" {
	fmt.Print("\033[38;5;172m")}

	var csc = client.InitAPIClient()
	fmt.Print("You are now con-nec-ted\n")
	time.Sleep(500 * time.Millisecond)
	//fmt.Print(csc)
	csc.Config = client.GetConfig()
	//fmt.Print(csc.Config)

	//cfg := Config{apiUrl: "https://api.cyberspace.online/v1"}
	//client := http.NewClientHandler()
	if csc.Config.StayLoggedIn == true && csc.Config.StoredValues.RefreshToken != "" {
		csc.Tokens = client.AuthTokens{RefreshToken: "", IDToken: "", RTDBToken: ""}
		csc.Tokens.RefreshToken = csc.Config.StoredValues.RefreshToken
		//fmt.Print((csc.Tokens.RefreshToken), "\n")
		csc.TokenRefresh()
		fmt.Print("You are still logged in.\n")

	} else {
		csc.Tokens = client.Login(csc.ApiUrl)
	}
	time.Sleep(500 * time.Millisecond)
	user, err := csc.GetMyUserProfile()
	if err != nil {
		fmt.Print(err)
	}
	csc.Username = user.Username

	fmt.Printf("Welcome to Cyberspace, @%s\n", csc.Username)
	time.Sleep(500 * time.Millisecond)
	//fmt.Printf("[authToken: %.10s...]\n", csc.Tokens.IDToken)

	c := commands{make(map[string]func(*client.APIClient, command) error)}
	c.register("view", handlerView)
	c.register("write", handlerWrite)
	c.register("edit", handlerEdit)
	c.register("publish", handlerPublish)
	c.register("bookmark", handlerBookmark)
	c.register("help", handlerHelp)
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
		if operatingSystem != "windows" {
		fmt.Print("\033[38;5;172m")
		if operatingSystem != "windows" {
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
	case "bookmarks":
		return handlerViewBookmarks(csc, cmd)
	case "profile":
		return handlerViewProfile(csc, cmd)
	case "user":
		return handlerViewProfile(csc, cmd)
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
	case "reply":
		return handlerWriteReply(csc, cmd)
	case "response":
		return handlerWriteReply(csc, cmd) //Because i mixed those up more than once.

	default:
		return fmt.Errorf("Unknown argument. Valid arguments for write: post, note.\n")

	}

}

func handlerEdit(csc *client.APIClient, cmd command) error {
	if len(cmd.Args) == 0 {
		renderPrint("The 'edit' command requires an argument. Valid arguments: note <note_id>, config.\n")
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

func handlerBookmark(csc *client.APIClient, cmd command) error {
	if len(cmd.Args) == 0 {
		renderPrint("The 'bookmark' command requires an argument: The id of the target.\n")
		return nil
	}

	targetSimpleID := cmd.Args[0]
	targetFullID, err := getFullID(targetSimpleID)
	fmt.Print(targetFullID)
	if err != nil {
		fmt.Print(err)
	}
	targetType := "post"
	if _, ok := csc.PostCache[targetFullID]; !ok {
		targetType = "reply"

	}
	err = csc.CreateBookmark(targetFullID, targetType)
	if err != nil {
		return fmt.Errorf("Failed to create bookmark: %s", err)
	}
	fmt.Print("Bookmark successfully created")
	return nil

}

func handlerHelp(csc *client.APIClient, cmd command) error {
	fmt.Print(`

 CyberspaceCLI supports the following commands: 

- view feed (optional_arg): Load 10 posts from the feed, starting at the newest. Every time the command is used, 10 more are loaded starting from where the previous iteration stopped. In the feed, posts are truncated at 1000 characters. To see the whole post, use the 'view post' command. 
  - Use the optional argument 'new' to load posts made since you started the client without losing the marker of the basic command. 
  - Use 'reset' to start over entirely. 
- view post <post_id>: This command shows the post specified by the id argument, plus the first 20 comments.
- view notifications (optional_arg): Load 10 notifications. If the notification is for a post or reply, you can use the shown id to open that post. 
  - Supports the same optional arguments as 'view feed'
- view notes: Loads 10 notes from your journal.
- view bookmarks: Load 10 bookmarks. Due to current API limitations, only bookmarked posts can be displayed, but not bookmarked replies. 
- view profile <username>: Displays a simplified version of that users profile, as well as their pinned post if they have one. Use 'me' as the username to see your own profile.
- write post: Opens your default text editor (or if you have non, nano (use ctrl+s, ctrl+x to exit)) and lets you write a post. Be aware that it might fail to post, so don't invest too much effort into it without copying the contents elsewhere before saving and closing the editor. After closing the editor, you'll have a chance to choose topics for the post.
- write reply <target_id>: Write a reply to the post or reply whose id you gave. Will ask for final confirmation before posting. 
- write note: Same as 'write post', but your writing is put in your journal instead.
- edit note <note_id: Opens a note in your default text editor (if none, nano) and lets you edit it.
- post <note_id>: Posts a note to the feed, making it visible to other users. 
- edit config: This lets you edit the client's config file. If you set 'stay logged in' to true, the client will save your refresh token and you will remain logged in across sessions. The config file should be in your .config/ or Library/Application Support/ directories, depending on whether you use linux or apple.
- bookmark <target_id: Bookmarks the post or reply whose id was given as an argument.
- help: you are >here<
- exit: exit	

`, "\n")
	return nil
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
				_, old_posts = getSimpleID(post.PostID) //Checks if this iteration crossed into new posts.
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
	bookmarks, _, err := csc.GetBookmarks(10, csc.Cursors["bookmarks"]) //Normal feed viewing.
	if err != nil {
		return err
	}

	for _, bookmark := range bookmarks {
		//bookmarkId, _ := getSimpleID(bookmark.BookmarkID)
		if bookmark.Type == "post" {
			if post, ok := csc.PostCache[bookmark.PostID]; ok {
				renderPost(post, true)
			} else {
				post, err := csc.GetPostById(bookmark.PostID)
				if err != nil {
					fmt.Print("Error while trying to retrieve bookmark post by id: ", err)
				}
				renderPost(post, true)
			}
		}
	}
	return nil

} // Limited function due to inability to target specific replies for retrieval.

func handlerViewProfile(csc *client.APIClient, cmd command) error {
	if len(cmd.Args) != 2 {
		fmt.Print("This command requires a target. Either a username or 'me' for your own profile.")
		return nil
	}
	target := cmd.Args[1]
	if target[0] == '@' {
		target = string(target[1:])
	}
	var user client.User
	var err error
	if target == "me" {
		user, err = csc.GetMyUserProfile()
	} else {
		user, err = csc.GetUserProfileByName(target)
	}
	if err != nil {
		return err
	}
	renderProfile(user)
	if user.PinnedPostID != "" {
		pinnedPost, err := csc.GetPostById(user.PinnedPostID)
		if err != nil {
			return err
		}
		renderPost(pinnedPost, true)
	}

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

func handlerWriteReply(csc *client.APIClient, cmd command) error {
	if len(cmd.Args) != 2 {
		renderPrint("The 'write reply' command requires the id of a target as an argument\n")
	}

	replyInput := client.CreateReplyInput{}

	targetSimpleID := cmd.Args[1]
	targetFullID, err := getFullID(targetSimpleID)
	//fmt.Print(targetFullID)
	if err != nil {
		fmt.Print(err)
	}
	//targetIsPost := false
	if _, ok := csc.PostCache[targetFullID]; !ok {
		replyInput.ParentReplyID = targetFullID
		replyInput.PostID = csc.ReplyCache[targetFullID].PostID
		//fmt.Printf("PostID: %s", replyInput.PostID)

	} else {
		replyInput.PostID = targetFullID
		//fmt.Printf("PostID: %s", replyInput.PostID)
	}

	reply, err := csc.CreateReply(replyInput)
	if err != nil {
		fmt.Print(err)
	}
	if reply.Content == "" { //If the user either wrote nothing in the document, or didn't confirm intention to post.
		return nil
	}
	renderReply(reply)
	return nil

}

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

func getSimpleID(fullID string) (simpleID int, exists bool) {
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
