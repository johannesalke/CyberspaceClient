package tui

import (
	"fmt"
	"testing"
)

func fillTracks(n int) []Track {
	tracks := make([]Track, 0, n)
	for i := 0; i < n; i++ {
		tracks = append(tracks, Track{Source: fmt.Sprintf("https://example.com/%d", i)})
	}
	return tracks
}

func TestJukeboxLoadMoreMergesAndTracksCursor(t *testing.T) {
	m := newTestModel()
	m.tracks = []Track{{Source: "https://example.com/a"}}
	m.tracksCursor = "abc"

	more := []Track{
		{Source: "https://example.com/a"},
		{Source: "https://example.com/b"},
	}
	updated, _ := m.Update(jukeboxLoadedMsg{tracks: more, cursor: "def", reset: false})
	m = updated.(Model)

	if len(m.tracks) != 2 {
		t.Fatalf("expected 2 tracks after load-more merge, got %d", len(m.tracks))
	}
	if m.tracksCursor != "def" {
		t.Fatalf("expected cursor to advance to def, got %q", m.tracksCursor)
	}
}

func TestJukeboxResetReplacesCatalogue(t *testing.T) {
	m := newTestModel()
	m.tracks = []Track{{Source: "https://example.com/old"}}
	m.tracksCursor = "zzz"
	m.jukeboxPage = 5

	updated, _ := m.Update(jukeboxLoadedMsg{tracks: []Track{{Source: "https://example.com/new"}}, cursor: "abc", reset: true})
	m = updated.(Model)

	if len(m.tracks) != 1 || m.tracks[0].Source != "https://example.com/new" {
		t.Fatalf("reset should replace the catalogue, got %+v", m.tracks)
	}
	if m.tracksCursor != "abc" {
		t.Fatalf("expected cursor abc, got %q", m.tracksCursor)
	}
	if m.jukeboxPage != 1 {
		t.Fatalf("reset should return to page 1, got %d", m.jukeboxPage)
	}
}

func TestJukeboxPageNextWithinLoadedCatalogue(t *testing.T) {
	m := newTestModel()
	m.page = jukeboxPage
	m.loading = false
	m.tracks = fillTracks(25)
	m.jukeboxPage = 1
	m.jukeboxIdx = 0

	updated, cmd := m.Update(press('n'))
	m = updated.(Model)
	if cmd != nil {
		t.Fatal("advancing within the loaded catalogue should not fetch")
	}
	if m.jukeboxPage != 2 || m.jukeboxIdx != 10 {
		t.Fatalf("expected page 2 / index 10, got page %d / index %d", m.jukeboxPage, m.jukeboxIdx)
	}
}

func TestJukeboxPageNextAtLastPageFetchesMore(t *testing.T) {
	m := newTestModel()
	m.page = jukeboxPage
	m.loading = false
	m.tracks = fillTracks(25)
	m.jukeboxPage = m.jukeboxLastPage()
	m.jukeboxIdx = m.jukeboxPageStart()
	m.tracksCursor = "abc"

	updated, cmd := m.Update(press('n'))
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("next page at the last loaded page should fetch more")
	}
	if !m.loading || !m.jukeboxAdvancePage {
		t.Fatal("expected loading and the advance-page flag to be set")
	}
}

func TestJukeboxPageNextAtLastPageInertWhenFeedExhausted(t *testing.T) {
	m := newTestModel()
	m.page = jukeboxPage
	m.loading = false
	m.tracks = fillTracks(25)
	m.jukeboxPage = m.jukeboxLastPage()
	m.tracksCursor = ""

	_, cmd := m.Update(press('n'))
	if cmd != nil {
		t.Fatal("next page should be inert once the feed is exhausted")
	}
}

func TestJukeboxPagePrevious(t *testing.T) {
	m := newTestModel()
	m.page = jukeboxPage
	m.tracks = fillTracks(25)
	m.jukeboxPage = 3
	m.jukeboxIdx = 20

	updated, cmd := m.Update(pressPageUp())
	m = updated.(Model)
	if cmd != nil {
		t.Fatal("previous page should not fetch")
	}
	if m.jukeboxPage != 2 || m.jukeboxIdx != 10 {
		t.Fatalf("expected page 2 / index 10, got page %d / index %d", m.jukeboxPage, m.jukeboxIdx)
	}
}

func TestJukeboxPagePreviousAtFirstPageInert(t *testing.T) {
	m := newTestModel()
	m.page = jukeboxPage
	m.tracks = fillTracks(5)
	m.jukeboxPage = 1

	_, cmd := m.Update(pressPageUp())
	if cmd != nil {
		t.Fatal("previous page at the first page should be inert")
	}
}
