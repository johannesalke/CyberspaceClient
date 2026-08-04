package main

import (
	"slices"
	"strings"
	"testing"

	client "github.com/johannesalke/cyberspacecli/internal/cyberspaceClient"
)

func TestParseInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"plain command", "help", []string{"help"}},
		{"slash command", "/help", []string{"help"}},
		{"slash command with arguments", "/view post 3", []string{"view", "post", "3"}},
		{"uppercase is normalised", "/HELP", []string{"help"}},
		{"only the command word is lowercased", "view profile Chovy", []string{"view", "profile", "Chovy"}},
		{"runs of spaces collapse", "view   post    3", []string{"view", "post", "3"}},
		{"surrounding whitespace is trimmed", "  \thelp  ", []string{"help"}},
		{"empty line", "", nil},
		{"whitespace only", "   ", nil},
		{"bare slash", "/", nil},
		{"repeated slashes", "//view feed", []string{"view", "feed"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseInput(tt.input)
			if !slices.Equal(got, tt.want) {
				t.Errorf("parseInput(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsExitWord(t *testing.T) {
	for _, input := range []string{"exit", "quit", "/exit", "/quit", "EXIT", " exit ", "Quit"} {
		if !isExitWord(input) {
			t.Errorf("isExitWord(%q) = false, want true", input)
		}
	}
	for _, input := range []string{"", "/", "exits", "view feed", "AutoResize=true"} {
		if isExitWord(input) {
			t.Errorf("isExitWord(%q) = true, want false", input)
		}
	}
}

// Every command the help text and README advertise has to resolve, whether or
// not the user typed a slash.
func TestRegisteredCommandsResolve(t *testing.T) {
	c := newCommands()

	for _, name := range []string{"view", "write", "edit", "publish", "post", "bookmark", "help", "delete", "logout"} {
		if _, ok := c.commands[name]; !ok {
			t.Errorf("command %q is not registered", name)
			continue
		}
		if _, ok := c.commands[parseInput("/" + name)[0]]; !ok {
			t.Errorf("command %q does not resolve when typed with a leading slash", name)
		}
	}
}

func TestUnknownCommandListsTheKnownOnes(t *testing.T) {
	c := newCommands()

	err := c.run(&client.APIClient{}, command{Name: "frobnicate"})
	if err == nil {
		t.Fatal("running an unregistered command returned no error")
	}
	for _, want := range []string{"frobnicate", "view", "help"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}
