package tui

import (
	"os"
	"testing"

	client "github.com/johannesalke/cyberspacecli/internal/cyberspaceClient"
)

func TestJukeboxNextPageLive(t *testing.T) {
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
	m.loading = false

	updated, cmd := m.switchPage(jukeboxPage)
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("switchPage(jukeboxPage) returned no command")
	}
	updated, _ = m.Update(cmd())
	m = updated.(Model)
	if m.err != nil {
		t.Fatalf("jukebox load error: %v", m.err)
	}
	first := len(m.tracks)
	t.Logf("after initial load: %d tracks, page %d/%d, cursor=%q", first, m.jukeboxPage, m.jukeboxLastPage(), m.tracksCursor)
	if first == 0 {
		t.Fatal("no tracks after initial load")
	}
	if m.tracksCursor == "" {
		t.Fatal("expected a resume cursor after initial load")
	}

	// Jump to the last loaded page and press n to fetch deeper.
	m.jukeboxPage = m.jukeboxLastPage()
	m.jukeboxIdx = m.jukeboxPageStart()

	updated, cmd = m.Update(press('n'))
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("n at the last loaded page did not issue a fetch command")
	}
	updated, _ = m.Update(cmd())
	m = updated.(Model)
	if m.err != nil {
		t.Fatalf("next page error: %v", m.err)
	}
	second := len(m.tracks)
	t.Logf("after next page: %d tracks, page %d/%d, cursor=%q", second, m.jukeboxPage, m.jukeboxLastPage(), m.tracksCursor)
	if second <= first {
		t.Fatalf("next page did not add tracks: first=%d second=%d", first, second)
	}
	if m.jukeboxPage != m.jukeboxLastPage() {
		t.Fatalf("expected to land on the last page after fetching, got page %d/%d", m.jukeboxPage, m.jukeboxLastPage())
	}
}
