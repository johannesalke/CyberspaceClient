// Command cyberspace starts the full-screen Cyberspace terminal client.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	client "github.com/johannesalke/cyberspacecli/internal/cyberspaceClient"
	"github.com/johannesalke/cyberspacecli/internal/tui"
)

// version is stamped at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("cyberspace " + version)
		return
	}

	csc := client.InitAPIClient()
	csc.Config = client.GetConfig()

	if csc.Config.Settings.StayLoggedIn && csc.Config.StoredValues.RefreshToken != "" {
		csc.Tokens.RefreshToken = csc.Config.StoredValues.RefreshToken
		if err := csc.TokenRefresh(); err != nil {
			fmt.Fprintln(os.Stderr, "Could not restore your session:", err)
			os.Exit(1)
		}
	} else {
		csc.Tokens = client.Login(csc.ApiUrl)
	}

	user, err := csc.GetMyUserProfile()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not load your Cyberspace profile:", err)
		os.Exit(1)
	}
	csc.Username = user.Username

	program := tea.NewProgram(tui.New(&csc))
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Cyberspace TUI error:", err)
		os.Exit(1)
	}

	if csc.Config.Settings.StayLoggedIn {
		csc.Config.StoredValues.RefreshToken = csc.Tokens.RefreshToken
		if err := csc.SaveConfig(csc.Config); err != nil {
			fmt.Fprintln(os.Stderr, "Could not save the session:", err)
		}
	}
}
