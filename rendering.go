package main

import (
	"fmt"
	"strings"
	"time"

	glamour "charm.land/glamour/v2"
	lipgloss "charm.land/lipgloss/v2"
	humanize "github.com/dustin/go-humanize"
	client "github.com/johannesalke/CyberspaceClient/internal/cyberspaceClient"
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

func renderPost(post client.Post, fullPost bool) {
	topline, _ := renderer.Render(fmt.Sprintln("@"+post.AuthorUsername, " | ", post.RepliesCount, " replies | ", post.PostID))

	seperator, err := renderer.Render(strings.Repeat("─", 80))
	if err != nil {
		fmt.Println(err)
	}
	renderedMD, err := renderer.Render(post.Content)
	if err != nil {
		fmt.Println(err)
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
	responseTarget := reply.ParentPostAuthor
	if reply.ParentReplyAuthor != "" {
		responseTarget = reply.ParentReplyAuthor
	}
	topline, _ := renderer.Render(fmt.Sprintln("@"+reply.AuthorUsername, " | ", "Responding to @"+responseTarget, " | ", reply.ReplyID))

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

func renderNotification(csc *client.APIClient, n client.Notification) {

	//timeSince :=time.Since(n.CreatedAt)
	timeSince := humanize.RelTime(time.Now(), n.CreatedAt, "in the future", "ago")
	var id = ""
	if n.Type == "new_post_friend" || n.Type == "new_post_following" || n.Type == "reply" {
		id = "| Id: " + n.TargetID
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
