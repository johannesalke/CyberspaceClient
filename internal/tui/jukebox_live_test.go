package tui

import (
	"os"
	"testing"

	client "github.com/johannesalke/cyberspacecli/internal/cyberspaceClient"
)

func TestJukeboxTracksFromLiveFeed(t *testing.T) {
	refresh := os.Getenv("CYBERSPACE_REFRESH_TOKEN")
	if refresh == "" {
		t.Skip("CYBERSPACE_REFRESH_TOKEN not set")
	}
	c := client.InitAPIClient()
	c.Tokens.RefreshToken = refresh
	if err := c.TokenRefresh(); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	m := New(&c)

	// Switching to the jukebox page runs the deeper feed scan that builds the
	// track catalogue.
	updated, cmd := m.switchPage(jukeboxPage)
	if cmd == nil {
		t.Fatal("switchPage(jukeboxPage) returned no command")
	}
	m = updated.(Model)
	updated, _ = m.Update(cmd())
	m = updated.(Model)
	if m.err != nil {
		t.Fatalf("jukebox load error: %v", m.err)
	}
	if len(m.tracks) == 0 {
		t.Fatal("jukebox found no audio tracks in the live feed")
	}
	t.Logf("tracks=%d", len(m.tracks))
	for _, tr := range m.tracks {
		t.Logf("track: %s | %s (%s)", tr.Artist, tr.Title, tr.Source)
	}
}
