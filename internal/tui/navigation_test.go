package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	client "github.com/johannesalke/cyberspacecli/internal/cyberspaceClient"
)

func press(code rune) tea.KeyMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}

func pressShiftTab() tea.KeyMsg {
	return tea.KeyPressMsg(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift})
}

func pressEsc() tea.KeyMsg {
	return tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc})
}

func pressPageUp() tea.KeyMsg {
	return tea.KeyPressMsg(tea.Key{Code: tea.KeyPgUp})
}

func newTestModel() Model {
	return New(&client.APIClient{})
}

func TestJukeboxBackReturnsToFeed(t *testing.T) {
	m := newTestModel()
	m.page = jukeboxPage
	m.tracks = []Track{{Source: "https://example.com/a"}}
	m.jukeboxIdx = 0

	updated, _ := m.Update(press(tea.KeyEsc))
	m = updated.(Model)
	if m.page != feedPage {
		t.Fatalf("esc on jukebox page: expected feedPage, got %q", m.page)
	}
}

func TestReplySelectionCycles(t *testing.T) {
	m := newTestModel()
	m.viewingPost = true
	m.activePost = client.Post{PostID: "root", AuthorUsername: "rootuser"}
	m.replies = []client.Reply{
		{ReplyID: "r1", AuthorUsername: "alice"},
		{ReplyID: "r2", AuthorUsername: "bob"},
	}
	m.replyIdx = -1

	tabs := []tea.KeyMsg{press(tea.KeyTab), press(tea.KeyTab), pressShiftTab(), pressShiftTab()}
	want := []int{0, 1, 0, -1}
	for i, key := range tabs {
		updated, _ := m.Update(key)
		m = updated.(Model)
		if m.replyIdx != want[i] {
			t.Fatalf("step %d: expected replyIdx %d, got %d", i, want[i], m.replyIdx)
		}
	}
}

func TestEnterOnSelectedReplyTargetsThatReply(t *testing.T) {
	m := newTestModel()
	m.viewingPost = true
	m.activePost = client.Post{PostID: "root", AuthorUsername: "rootuser"}
	m.replies = []client.Reply{
		{ReplyID: "r1", AuthorUsername: "alice"},
		{ReplyID: "r2", AuthorUsername: "bob"},
	}
	m.replyIdx = 1

	updated, _ := m.Update(press(tea.KeyEnter))
	m = updated.(Model)
	if !m.composing || !m.replying {
		t.Fatal("enter on selected reply should open the reply composer")
	}
	if m.replyParentID != "r2" {
		t.Fatalf("expected parent reply r2, got %q", m.replyParentID)
	}
	if m.replyParentAuthor != "bob" {
		t.Fatalf("expected parent author bob, got %q", m.replyParentAuthor)
	}
}

func TestReplyToRootPostHasNoParent(t *testing.T) {
	m := newTestModel()
	m.viewingPost = true
	m.activePost = client.Post{PostID: "root", AuthorUsername: "rootuser"}
	m.replies = []client.Reply{{ReplyID: "r1", AuthorUsername: "alice"}}
	m.replyIdx = -1

	updated, _ := m.Update(press('r'))
	m = updated.(Model)
	if !m.composing || !m.replying {
		t.Fatal("r on post detail should open the reply composer")
	}
	if m.replyParentID != "" {
		t.Fatalf("replying to root post should have no parent reply, got %q", m.replyParentID)
	}
	if m.replyParentAuthor != "rootuser" {
		t.Fatalf("expected root author, got %q", m.replyParentAuthor)
	}
}
