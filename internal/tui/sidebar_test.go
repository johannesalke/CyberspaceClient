package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	client "github.com/johannesalke/cyberspacecli/internal/cyberspaceClient"
)

func TestSidebarWidthThreshold(t *testing.T) {
	hidden := newTestModel()
	hidden.width = 79
	if w := hidden.sidebarWidth(); w != 0 {
		t.Fatalf("very narrow terminal should hide the sidebar, got width %d", w)
	}
	compact := newTestModel()
	compact.width = 80
	if w := compact.sidebarWidth(); w != 26 {
		t.Fatalf("80-column terminal should show a 26-column sidebar, got %d", w)
	}
	medium := newTestModel()
	medium.width = 100
	if w := medium.sidebarWidth(); w != 30 {
		t.Fatalf("100-column terminal should show a 30-column sidebar, got %d", w)
	}
	wide := newTestModel()
	wide.width = 120
	if w := wide.sidebarWidth(); w != 34 {
		t.Fatalf("wide terminal should show a 34-column sidebar, got %d", w)
	}
}

func TestShiftedNMatchesFocusBinding(t *testing.T) {
	m := newTestModel()
	m.width = 120
	shifted := tea.KeyPressMsg(tea.Key{Code: 'n', Mod: tea.ModShift, Text: "N"})
	if !m.matches("focus_notifications", shifted) {
		t.Fatal("shift+n (as a real terminal delivers it) should match the focus_notifications binding")
	}
	updated, cmd := m.Update(shifted)
	m = updated.(Model)
	if m.sidebarFocus != sidebarFocusNotifications {
		t.Fatal("shift+n should focus the sidebar notifications panel")
	}
	if cmd != nil {
		t.Fatalf("focusing the sidebar should not load anything, got cmd %v", cmd)
	}
}

func TestPlainNMatchesFocusBinding(t *testing.T) {
	m := newTestModel()
	m.width = 120
	if !m.matches("focus_notifications", press('n')) {
		t.Fatal("plain n should also match the focus_notifications binding")
	}
	updated, cmd := m.Update(press('n'))
	m = updated.(Model)
	if m.sidebarFocus != sidebarFocusNotifications {
		t.Fatal("plain n should focus the sidebar notifications panel")
	}
	if cmd != nil {
		t.Fatalf("focusing the sidebar should not load anything, got cmd %v", cmd)
	}
}

func TestOlderPostsMovedOffN(t *testing.T) {
	m := newTestModel()
	m.width, m.height = 120, 30
	m.cursor = "cursor-1"
	if m.matches("next_page", press('n')) {
		t.Fatal("n should no longer load older posts, it is the notifications shortcut now")
	}
	if !m.matches("next_page", press('o')) {
		t.Fatal("plain o should load older posts")
	}
	if !m.matches("next_page", tea.KeyPressMsg(tea.Key{Code: 'o', Mod: tea.ModShift, Text: "O"})) {
		t.Fatal("shift+o should also load older posts")
	}
}

func TestFocusNotificationsFallsBackToPage(t *testing.T) {
	m := newTestModel()
	m.width = 79
	m.page = feedPage
	updated, cmd := m.Update(press('N'))
	m = updated.(Model)
	if m.page != notificationsPage {
		t.Fatalf("N on a terminal with no sidebar should open the notifications page, got %q", m.page)
	}
	if cmd == nil {
		t.Fatal("opening the notifications page should kick off a load")
	}
}

func TestBackExitsNotificationsPage(t *testing.T) {
	m := newTestModel()
	m.width, m.height = 120, 30
	m.page = notificationsPage
	m.err = fmt.Errorf("stale error")
	updated, cmd := m.Update(pressEsc())
	m = updated.(Model)
	if m.page != feedPage {
		t.Fatalf("esc should leave the notifications page and return to the feed, got %q", m.page)
	}
	if m.sidebarFocus != sidebarFocusFeed {
		t.Fatal("leaving the notifications page should reset sidebar focus to the feed")
	}
	if m.err != nil {
		t.Fatalf("leaving the notifications page should clear errors, got %v", m.err)
	}
	if cmd != nil {
		t.Fatalf("leaving the notifications page should not load anything, got cmd %v", cmd)
	}
}

func TestNotificationsFooterShowsBack(t *testing.T) {
	m := newTestModel()
	m.page = notificationsPage
	if line := m.helpLine(); !strings.Contains(line, m.keyNames("back")+" back") {
		t.Fatalf("the notifications page footer should advertise esc back, got %q", line)
	}
	feed := newTestModel()
	feed.page = feedPage
	if line := feed.helpLine(); strings.Contains(line, m.keyNames("back")+" back") {
		t.Fatalf("the feed page footer should not advertise back, got %q", line)
	}
}

func TestFeedSelectionRevealsPost(t *testing.T) {
	m := newTestModel()
	m.width, m.height = 120, 30
	m.resizeViewport()

	var posts []client.Post
	for i := 0; i < 20; i++ {
		posts = append(posts, client.Post{
			PostID:         fmt.Sprintf("p%d", i),
			AuthorUsername: "user",
			Content:        strings.Repeat("word ", 200),
		})
	}
	m.posts = posts
	m.selectedPost = len(posts) - 1
	m.renderFeed()

	line := m.feedOffsets[m.selectedPost]
	if line < 0 {
		t.Fatal("selected post was not rendered")
	}
	m.revealFeedSelection()

	top := m.viewport.YOffset()
	bottom := top + m.viewport.Height()
	if line < top || line >= bottom {
		t.Fatalf("selected post line %d not visible in window [%d, %d)", line, top, bottom)
	}
}

func TestTabThroughFeedKeepsSelectionVisible(t *testing.T) {
	m := newTestModel()
	m.width, m.height = 120, 20
	m.resizeViewport()

	var posts []client.Post
	for i := 0; i < 12; i++ {
		posts = append(posts, client.Post{
			PostID:         fmt.Sprintf("p%d", i),
			AuthorUsername: "user",
			Content:        strings.Repeat("longword ", 60),
		})
	}
	m.posts = posts
	m.renderFeed()

	for i := 0; i < len(posts); i++ {
		updated, _ := m.Update(press(tea.KeyTab))
		m = updated.(Model)
		line := m.feedOffsets[m.selectedPost]
		top := m.viewport.YOffset()
		bottom := top + m.viewport.Height()
		if line < top || line >= bottom {
			t.Fatalf("step %d: selected line %d outside window [%d, %d)", i, line, top, bottom)
		}
	}
}

func TestFocusNotificationsCyclesAndEscapes(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.notifications = []client.Notification{
		{ID: "n1", ActorUsername: "alice", Type: "reply"},
		{ID: "n2", ActorUsername: "bob", Type: "new_follower"},
		{ID: "n3", ActorUsername: "carol", Type: "poke"},
	}

	updated, _ := m.Update(press('N'))
	m = updated.(Model)
	if m.sidebarFocus != sidebarFocusNotifications {
		t.Fatal("N should focus the notifications panel")
	}

	updated, _ = m.Update(press(tea.KeyTab))
	m = updated.(Model)
	updated, _ = m.Update(press(tea.KeyTab))
	m = updated.(Model)
	if m.notifIdx != 2 {
		t.Fatalf("expected selection at index 2, got %d", m.notifIdx)
	}

	updated, _ = m.Update(press(tea.KeyTab))
	m = updated.(Model)
	if m.notifIdx != 2 {
		t.Fatalf("selection should clamp at the last notification, got %d", m.notifIdx)
	}

	updated, _ = m.Update(pressShiftTab())
	m = updated.(Model)
	if m.notifIdx != 1 {
		t.Fatalf("expected selection back at index 1, got %d", m.notifIdx)
	}

	updated, _ = m.Update(press(tea.KeyEsc))
	m = updated.(Model)
	if m.sidebarFocus != sidebarFocusFeed {
		t.Fatal("esc should return focus to the feed")
	}
}

func TestOpenPostNotificationLoadsDetail(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.sidebarFocus = sidebarFocusNotifications
	m.notifications = []client.Notification{
		{ID: "n1", ActorUsername: "alice", Type: "new_post_following", TargetID: "post-1"},
	}
	m.notifIdx = 0

	updated, cmd := m.Update(press(tea.KeyEnter))
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("opening a post notification should return a load command")
	}
	if !m.loading {
		t.Fatal("expected the post detail load to begin")
	}
	if !m.notifications[0].Read {
		t.Fatal("opened notification should be marked read")
	}
	if m.pendingReplyID != "" {
		t.Fatalf("a post notification should not set a pending reply, got %q", m.pendingReplyID)
	}
}

func TestOpenReplyNotificationHighlightsReply(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.sidebarFocus = sidebarFocusNotifications
	m.notifications = []client.Notification{
		{ID: "n1", ActorUsername: "alice", Type: "reply", TargetID: "post-1"},
	}
	m.notifications[0].Metadata.ReplyID = "r7"
	m.notifIdx = 0

	updated, _ := m.Update(press(tea.KeyEnter))
	m = updated.(Model)
	if m.pendingReplyID != "r7" {
		t.Fatalf("expected pending reply r7, got %q", m.pendingReplyID)
	}

	updated, _ = m.Update(postDetailLoadedMsg{
		post:    client.Post{PostID: "post-1", AuthorUsername: "root"},
		replies: []client.Reply{{ReplyID: "r7", AuthorUsername: "alice"}, {ReplyID: "r8", AuthorUsername: "dave"}},
	})
	m = updated.(Model)
	if m.replyIdx != 0 {
		t.Fatalf("expected the pending reply to be selected, got index %d", m.replyIdx)
	}
	if m.pendingReplyID != "" {
		t.Fatalf("pending reply should clear after the detail loads, got %q", m.pendingReplyID)
	}
}

func TestOpenSocialNotificationShowsNotice(t *testing.T) {
	m := newTestModel()
	m.width = 120
	m.sidebarFocus = sidebarFocusNotifications
	m.notifications = []client.Notification{
		{ID: "n1", ActorUsername: "bob", Type: "new_follower"},
	}
	m.notifIdx = 0

	updated, cmd := m.Update(press(tea.KeyEnter))
	m = updated.(Model)
	if m.loading {
		t.Fatal("a social notification should not start a page load")
	}
	if m.notice == "" {
		t.Fatal("expected a notice explaining there is nothing to view")
	}
	if cmd == nil {
		t.Fatal("a social notification should still mark itself read")
	}
}

func TestGlobalPlayerControlsOnFeed(t *testing.T) {
	m := newTestModel()
	m.width, m.height = 120, 40
	m.page = feedPage
	m.tracks = fillTracks(2)
	m.nowPlaying = 0

	updated, cmd := m.Update(press('p'))
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("pause should work globally while a track is playing")
	}

	updated, cmd = m.Update(press(tea.KeyLeft))
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("previous should work globally while a track is playing")
	}
	if m.jukeboxIdx != 1 {
		t.Fatalf("previous should wrap to the last track, got index %d", m.jukeboxIdx)
	}
}

func TestGlobalJukeboxNextRespectsFeedCursor(t *testing.T) {
	// With no older posts left to load, right advances the track.
	m := newTestModel()
	m.width, m.height = 120, 40
	m.page = feedPage
	m.tracks = fillTracks(2)
	m.nowPlaying = 0
	m.cursor = ""

	updated, cmd := m.Update(press(tea.KeyRight))
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("right with an exhausted feed should advance the track")
	}
	if m.jukeboxIdx != 1 {
		t.Fatalf("expected the next track, got index %d", m.jukeboxIdx)
	}

	// With older posts available, right keeps loading older posts instead.
	m2 := newTestModel()
	m2.width, m2.height = 120, 40
	m2.page = feedPage
	m2.loading = false
	m2.tracks = fillTracks(2)
	m2.nowPlaying = 0
	m2.cursor = "abc"

	_, cmd2 := m2.Update(press(tea.KeyRight))
	if cmd2 == nil {
		t.Fatal("right should keep loading older posts while a feed cursor exists")
	}
}

func TestGlobalPlayerKeysInertWithoutPlayingTrack(t *testing.T) {
	m := newTestModel()
	m.width, m.height = 120, 40
	m.page = feedPage
	m.nowPlaying = -1

	updated, cmd := m.Update(press('p'))
	m = updated.(Model)
	if cmd != nil {
		t.Fatal("player keys should be inert on the feed when nothing is playing")
	}
}

func TestSidebarViewRenders(t *testing.T) {
	m := newTestModel()
	m.width, m.height = 120, 40
	m.resizeViewport()
	m.notifications = []client.Notification{
		{ID: "n1", ActorUsername: "alice", Type: "reply"},
	}
	m.sidebarFocus = sidebarFocusNotifications
	m.notifIdx = 0

	if s := m.sidebarView(); s == "" {
		t.Fatal("sidebar should render its panels")
	}

	m.nowPlaying = 0
	m.tracks = fillTracks(1)
	if s := m.playerPanelContent(34); s == "" {
		t.Fatal("player panel should render a playing track")
	}
}

func TestSidebarHeightMatchesViewport(t *testing.T) {
	m := newTestModel()
	m.width, m.height = 120, 40
	m.resizeViewport()
	m.notifications = make([]client.Notification, 15)
	for i := range m.notifications {
		m.notifications[i] = client.Notification{ID: fmt.Sprintf("n%d", i), ActorUsername: "user", Type: "reply"}
	}
	m.nowPlaying = 0
	m.tracks = fillTracks(1)

	viewHeight := max(m.height-3, 1)
	sidebarHeight := strings.Count(m.sidebarView(), "\n") + 1
	if sidebarHeight > viewHeight {
		t.Fatalf("sidebar height %d exceeds viewport height %d", sidebarHeight, viewHeight)
	}
	if sidebarHeight < viewHeight {
		t.Fatalf("sidebar height %d does not fill viewport height %d", sidebarHeight, viewHeight)
	}
}

func TestNotificationTypeLabels(t *testing.T) {
	cases := map[string]string{
		"reply":              "replied",
		"thread_reply":       "replied",
		"new_follower":       "followed",
		"poke":               "poked",
		"new_post_following": "posted",
		"dm_message":         "message",
	}
	for raw, want := range cases {
		if got := notificationTypeLabel(raw); got != want {
			t.Fatalf("label for %q: expected %q, got %q", raw, want, got)
		}
	}
}
