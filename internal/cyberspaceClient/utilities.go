package client

import (
	//"github.com/johannesalke/CyberspaceTUI/internal/auth"
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func makeRequest(method, url string, tokens AuthTokens, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+tokens.IDToken)

	return req, nil
}

func makeGetUrl(url string, limit int, cursor string) string {
	if limit == 0 {
		limit = 20
	}

	url += fmt.Sprintf("?limit=%d", limit)
	if cursor != "" {
		url += fmt.Sprintf("&cursor=%s", cursor)
	}
	return url
}

func (c *APIClient) sendRequest(req *http.Request) (*http.Response, error) {
	res, err := c.Client.Do(req)
	c.LastStatusCode = res.StatusCode
	if err != nil {
		return res, err
	}
	return res, nil
}

/////////////////| Functions for writing Posts or Notes |//////////////////////

func WriteContent() string {
	tmpFile, err := os.CreateTemp("", "message-*.txt")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano" // fallback
	}

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		panic(err)
	}

	fmt.Println("\nMessage:")
	fmt.Println(string(content))
	return string(content)
}

func WriteTopics() []string {
	fmt.Print("Content registered. You may now add up to three topics to the entry, seperated by commas:\n")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	topicString := scanner.Text()
	topics := strings.Split(topicString, ",")
	return topics
}

func EditNote(note Note) (CreateNoteInput, error) {
	tmpFile, err := os.CreateTemp("", "message-*.txt")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(note.Content)

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vm" // fallback
	}

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		panic(err)
	}

	topics := note.Topics
	newNote := CreateNoteInput{Content: string(content), Topics: topics}

	return newNote, nil
}
