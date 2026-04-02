package main

import (
	"fmt"
	"strings"
	"time"

	glamour "charm.land/glamour/v2"
	lipgloss "charm.land/lipgloss/v2"
	humanize "github.com/dustin/go-humanize"
	client "github.com/johannesalke/cyberspacecli/internal/cyberspaceClient"
)

var (
	basicBox = lipgloss.NewStyle().
			Width(86).
			MarginLeft(4).
			Padding(0, 2, 0, 2).
			Foreground(lipgloss.Color("#ff9a10")).
			BorderForeground(lipgloss.Color("#744b0f"))

	boxTop = lipgloss.NewStyle().Inherit(basicBox).
		Border(lipgloss.RoundedBorder(), true, true, false, true).
		Padding(0, 2, 0, 2).
		MarginLeft(4).
		MarginTop(1)
	boxSides = lipgloss.NewStyle().Inherit(basicBox).
			Border(lipgloss.RoundedBorder(), false, true, false, true).
			Padding(0, 2, 0, 2).
			MarginLeft(4)
	boxBottom = lipgloss.NewStyle().Inherit(basicBox).
			Border(lipgloss.RoundedBorder(), false, true, true, true).
			Padding(0, 2, 0, 2).
			MarginLeft(4)
	thinBox = lipgloss.NewStyle().Inherit(basicBox).
		Border(lipgloss.RoundedBorder()).
		MarginLeft(4).
		Padding(0, 2, 0, 2)
)

var renderer, err = glamour.NewTermRenderer(
	glamour.WithStylePath("style.json"),
	glamour.WithWordWrap(80))

func RenderBox(elements ...string) error {
	N := len(elements)

	result := boxTop.Render(strings.TrimRight(elements[0], "\n")) + "\n"
	for _, element := range elements[1 : N-1] {
		result += boxSides.Render(strings.TrimRight(element, "\n")) + "\n"

	}
	result += boxBottom.Render(strings.TrimRight(elements[N-1], "\n")) + "\n"

	fmt.Print(result)
	return nil
}

func renderPost(post client.Post, fullPost bool) { //Full post should be set to false in the feed to truncate posts in the feed. THis is not implemented yet!

	simpleID, _ := simplifyID(post.PostID)
	topline, _ := renderer.Render(fmt.Sprintln("@"+post.AuthorUsername, " | ", post.RepliesCount, " replies | Id: ", simpleID))

	seperator, err := renderer.Render(strings.Repeat("─", 80))
	if err != nil {
		fmt.Println(err)
	}
	var renderedMD string

	if fullPost == false && len(post.Content) > 500 {

		//truncatedPost, _ := renderer.Render("...view post to continue")
		renderedMD, err = renderer.Render(fmt.Sprintf("%.1000s   %s", post.Content, "...*view post to continue*"))
		if err != nil {
			fmt.Println(err)
		}
	} else {

		renderedMD, err = renderer.Render(post.Content)
		if err != nil {
			fmt.Println(err)
		}
	}

	if len(post.Topics) != 0 { //This block occurs if a post has topic tags. In that case, another dividing line is added and the topics are displayed below it.

		topics, err := renderer.Render("Topics: " + strings.Join(post.Topics, ", "))
		if err != nil {
			fmt.Println(err)
		}
		err = RenderBox(topline, seperator, renderedMD, seperator, topics)
		if err != nil {
			fmt.Println(err)
		}
	} else { // Otherwise, just render normally.
		err = RenderBox(topline, seperator, renderedMD)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func renderReply(reply client.Reply) {
	simpleID, _ := simplifyID(reply.ReplyID)
	responseTarget := reply.ParentPostAuthor
	if reply.ParentReplyAuthor != "" {
		responseTarget = reply.ParentReplyAuthor
	}
	topline, _ := renderer.Render(fmt.Sprintln("@"+reply.AuthorUsername, " | ", "Responding to @"+responseTarget, " | Id: ", simpleID))

	seperator, err := renderer.Render(strings.Repeat("─", 80))
	if err != nil {
		fmt.Println(err)
	}
	renderedMD, err := renderer.Render(reply.Content)
	if err != nil {
		fmt.Println(err)
	}
	err = RenderBox(topline, seperator, renderedMD)
	if err != nil {
		fmt.Println(err)
	}

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

func renderNote(note client.Note, fullNote bool) { //Full note should be set to false in the feed to truncate posts in the feed. THis is not implemented yet!
	simpleID, _ := simplifyID(note.NoteID)
	topline, _ := renderer.Render(fmt.Sprintln("Id: ", simpleID))

	seperator, err := renderer.Render(strings.Repeat("─", 80))
	if err != nil {
		fmt.Println(err)
	}
	renderedMD, err := renderer.Render(note.Content)
	if err != nil {
		fmt.Println(err)
	}

	if len(note.Topics) != 0 { //This block occurs if a post has topic tags. In that case, another dividing line is added and the topics are displayed below it.

		topics, err := renderer.Render("Topics: " + strings.Join(note.Topics, ", "))
		if err != nil {
			fmt.Println(err)
		}
		err = RenderBox(topline, seperator, renderedMD, seperator, topics)
		if err != nil {
			fmt.Println(err)
		}
	} else { // Otherwise, just render normally.
		err = RenderBox(topline, seperator, renderedMD)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func renderNotification(csc *client.APIClient, n client.Notification) {
	simpleID, _ := simplifyID(n.TargetID)
	//timeSince :=time.Since(n.CreatedAt)
	timeSince := humanize.RelTime(time.Now(), n.CreatedAt, "in the future", "ago")
	var id = ""
	if n.Type == "new_post_friend" || n.Type == "new_post_following" || n.Type == "reply" {
		id = fmt.Sprintln("| Id: ", simpleID)
	}

	notification_string := fmt.Sprintf("User: %s | Type: %s | %s %s", n.ActorUsername, n.Type, timeSince, id)
	fmt.Print(thinBox.Render(notification_string) + "\n")
}

func renderText(str string) string {
	res, _ := renderer.Render(str)
	return res
}

func renderPrint(str string) {
	fmt.Print(renderer.Render(str))
}
