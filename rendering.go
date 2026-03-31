package main

import (
	glamour "charm.land/glamour/v2"
	lipgloss "charm.land/lipgloss/v2"
	"fmt"
	client "github.com/johannesalke/CyberspaceClient/internal/cyberspaceClient"
	"strings"
)

var (
	basicBox = lipgloss.NewStyle().
			Width(88).
			MarginLeft(2).
			Foreground(lipgloss.Color("#ff9a10")).
			BorderForeground(lipgloss.Color("#744b0f"))

	boxTop = lipgloss.NewStyle().Inherit(basicBox).
		Border(lipgloss.RoundedBorder(), true, true, false, true).
		Padding(0, 2, 0, 2).
		MarginLeft(4).
		MarginTop(2)
	boxSides = lipgloss.NewStyle().Inherit(basicBox).
			Border(lipgloss.RoundedBorder(), false, true, false, true).
			Padding(0, 2, 0, 2).
			MarginLeft(4)
	boxBottom = lipgloss.NewStyle().Inherit(basicBox).
			Border(lipgloss.RoundedBorder(), false, true, true, true).
			Padding(0, 2, 0, 2).
			MarginLeft(4)
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

func renderPost(post client.Post) {
	topline, _ := renderer.Render(fmt.Sprintln("@"+post.AuthorUsername, " | ", post.RepliesCount, " replies | ", post.PostID))

	seperator, err := renderer.Render(strings.Repeat("─", 80))
	if err != nil {
		fmt.Println(err)
	}
	renderedMD, err := renderer.Render(post.Content)
	if err != nil {
		fmt.Println(err)
	}
	err = RenderBox(topline, seperator, renderedMD)
	if err != nil {
		fmt.Println(err)
	}

}

func renderReplies(reply client.Reply) {
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

func renderText(str string) string {
	res, _ := renderer.Render(str)
	return res
}
