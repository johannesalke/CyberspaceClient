// Package tui contains the full-screen terminal interface for Cyberspace.
package tui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	client "github.com/johannesalke/cyberspacecli/internal/cyberspaceClient"
)

const (
	defaultTheme = "amber"
)

type theme struct {
	name   string
	accent string
	muted  string
	border string
}

var themes = []theme{
	{name: "amber", accent: "#ff9a10", muted: "#8c8c8c", border: "#744b0f"},
	{name: "phosphor", accent: "#8aff80", muted: "#79a873", border: "#356b35"},
	{name: "violet", accent: "#c7a6ff", muted: "#a190bc", border: "#65527c"},
	{name: "mono", accent: "#f0f0f0", muted: "#a0a0a0", border: "#707070"},
}

// wordmark is the client's stylized brand line, shown as the TUI headline.
const wordmark = "ᑕ¥βєяรקค¢є"

var (
	titleStyle lipgloss.Style
	metaStyle  lipgloss.Style
	postStyle  lipgloss.Style
	helpStyle  lipgloss.Style
)

type feedLoadedMsg struct {
	posts  []client.Post
	cursor string
	reset  bool
	err    error
}

type pageID string

const (
	feedPage          pageID = "feed"
	bookmarksPage     pageID = "bookmarks"
	notificationsPage pageID = "notifications"
	journalPage       pageID = "journal"
	profilePage       pageID = "profile"
	mailPage          pageID = "c-mail"
	jukeboxPage       pageID = "jukebox"
)

type notificationsLoadedMsg struct {
	items []client.Notification
	err   error
}

type bookmarksLoadedMsg struct {
	items []client.Bookmark
	err   error
}

type notesLoadedMsg struct {
	items []client.Note
	err   error
}

type profileLoadedMsg struct {
	user client.User
	err  error
}

type postCreatedMsg struct {
	post client.Post
	err  error
}

type postDetailLoadedMsg struct {
	post    client.Post
	replies []client.Reply
	err     error
}

type bookmarkCreatedMsg struct{ err error }

type replyCreatedMsg struct {
	reply client.Reply
	err   error
}

type playerStatusMsg struct {
	action string
	index  int
	paused bool
	err    error
}

type jukeboxLoadedMsg struct {
	tracks []Track
	cursor string
	reset  bool
	err    error
}

// Model displays a read-only, keyboard-navigable feed. Posting and other
// actions will be added as dedicated pages rather than command-line verbs.
type Model struct {
	client             *client.APIClient
	viewport           viewport.Model
	posts              []client.Post
	bookmarks          []client.Bookmark
	notifications      []client.Notification
	notes              []client.Note
	profile            client.User
	page               pageID
	cursor             string
	width              int
	height             int
	loading            bool
	err                error
	showHelp           bool
	keys               map[string][]string
	composer           textarea.Model
	composing          bool
	confirmPost        bool
	posting            bool
	replying           bool
	theme              string
	selectedPost       int
	viewingPost        bool
	activePost         client.Post
	replies            []client.Reply
	replyIdx           int
	replyParentID      string
	replyParentAuthor  string
	notice             string
	tracks             []Track
	tracksCursor       string
	player             *audioPlayer
	jukeboxIdx         int
	jukeboxPage        int
	jukeboxAdvancePage bool
	nowPlaying         int
	paused             bool
	playerErr          error
}

// New creates the Cyberspace feed TUI.
func New(c *client.APIClient) Model {
	pager := viewport.New()
	pager.KeyMap = disabledViewportKeyMap()
	selectedTheme := normalizeTheme(c.Config.Settings.Theme)
	applyTheme(selectedTheme)
	return Model{
		client:      c,
		loading:     true,
		viewport:    pager,
		keys:        client.ResolveKeyBindings(c.Config.Settings.KeyBindings),
		page:        feedPage,
		theme:       selectedTheme,
		player:      newAudioPlayer(),
		jukeboxPage: 1,
		nowPlaying:  -1,
	}
}

func normalizeTheme(name string) string {
	for _, candidate := range themes {
		if name == candidate.name {
			return name
		}
	}
	return defaultTheme
}

func applyTheme(name string) {
	selected := themes[0]
	for _, candidate := range themes {
		if candidate.name == name {
			selected = candidate
			break
		}
	}
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(selected.accent))
	metaStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(selected.muted))
	postStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(selected.border)).Padding(0, 1)
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(selected.muted))
}

// Bubble's pager defaults are disabled so that every keyboard action is
// governed by the user's config, including actions they intentionally disable.
func disabledViewportKeyMap() viewport.KeyMap {
	disabled := key.NewBinding(key.WithDisabled())
	return viewport.KeyMap{
		PageDown:     disabled,
		PageUp:       disabled,
		HalfPageUp:   disabled,
		HalfPageDown: disabled,
		Down:         disabled,
		Up:           disabled,
		Left:         disabled,
		Right:        disabled,
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadFeed("", true)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resizeViewport()
		if m.composing {
			m.composer.SetWidth(max(m.width-6, 30))
			m.composer.SetHeight(max(m.height-8, 6))
		}
		m.renderCurrentPage()
		return m, nil
	case feedLoadedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err != nil {
			return m, nil
		}
		if msg.reset {
			m.posts = msg.posts
			m.selectedPost = 0
		} else {
			m.posts = append(m.posts, msg.posts...)
		}
		m.cursor = msg.cursor
		m.tracks = mergeTracks(m.tracks, audioTracksFromPosts(msg.posts))
		m.renderCurrentPage()
		return m, nil
	case notificationsLoadedMsg:
		m.loading, m.err, m.notifications = false, msg.err, msg.items
		m.renderCurrentPage()
		return m, nil
	case bookmarksLoadedMsg:
		m.loading, m.err, m.bookmarks = false, msg.err, msg.items
		m.renderCurrentPage()
		return m, nil
	case notesLoadedMsg:
		m.loading, m.err, m.notes = false, msg.err, msg.items
		m.renderCurrentPage()
		return m, nil
	case profileLoadedMsg:
		m.loading, m.err, m.profile = false, msg.err, msg.user
		m.renderCurrentPage()
		return m, nil
	case pageLoadedPlaceholderMsg:
		m.loading = false
		m.renderCurrentPage()
		return m, nil
	case postCreatedMsg:
		m.posting = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.composing, m.confirmPost = false, false
		m.posts = append([]client.Post{msg.post}, m.posts...)
		m.page = feedPage
		m.renderFeed()
		return m, nil
	case postDetailLoadedMsg:
		m.loading, m.err = false, msg.err
		if msg.err == nil {
			m.activePost, m.replies, m.viewingPost = msg.post, msg.replies, true
			m.replyIdx = -1
			m.renderPostDetail()
		}
		return m, nil
	case bookmarkCreatedMsg:
		m.loading, m.err = false, msg.err
		if msg.err == nil {
			m.notice = "Bookmark saved."
		}
		return m, nil
	case replyCreatedMsg:
		m.posting = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.composing, m.confirmPost, m.replying = false, false, false
		m.replies = append(m.replies, msg.reply)
		m.renderPostDetail()
		return m, nil
	case playerStatusMsg:
		m.playerErr = msg.err
		switch msg.action {
		case "play":
			if msg.err == nil {
				m.nowPlaying, m.paused = msg.index, msg.paused
			}
		case "pause":
			if msg.err == nil {
				m.paused = msg.paused
			}
		case "stop":
			if msg.err == nil {
				m.nowPlaying = -1
			}
		}
		m.renderJukebox()
		return m, nil
	case jukeboxLoadedMsg:
		m.loading, m.err = false, msg.err
		if msg.err == nil {
			if msg.reset {
				m.tracks = msg.tracks
				m.jukeboxPage = 1
				m.jukeboxAdvancePage = false
			} else {
				m.tracks = mergeTracks(m.tracks, msg.tracks)
			}
			m.tracksCursor = msg.cursor
			if m.jukeboxAdvancePage {
				m.jukeboxPage = m.jukeboxLastPage()
				m.jukeboxAdvancePage = false
			}
			if m.jukeboxPage < 1 {
				m.jukeboxPage = 1
			}
			m.jukeboxIdx = (m.jukeboxPage - 1) * m.jukeboxPageSize()
			if m.nowPlaying >= len(m.tracks) {
				m.nowPlaying = -1
			}
			m.renderJukebox()
		}
		return m, nil
	case tea.KeyMsg:
		if m.composing {
			return m.updateComposer(msg)
		}
		if m.viewingPost {
			return m.updatePostDetail(msg)
		}
		switch {
		case m.matches("quit", msg):
			m.player.Stop()
			return m, tea.Quit
		case m.showHelp && m.matches("close_help", msg):
			m.showHelp = false
			return m, nil
		case m.matches("help", msg):
			m.showHelp = !m.showHelp
			return m, nil
		case !m.showHelp && m.matches("page_feed", msg):
			return m.switchPage(feedPage)
		case !m.showHelp && m.matches("page_bookmarks", msg):
			return m.switchPage(bookmarksPage)
		case !m.showHelp && m.matches("page_notifications", msg):
			return m.switchPage(notificationsPage)
		case !m.showHelp && m.matches("page_journal", msg):
			return m.switchPage(journalPage)
		case !m.showHelp && m.matches("page_profile", msg):
			return m.switchPage(profilePage)
		case !m.showHelp && m.matches("page_mail", msg):
			return m.switchPage(mailPage)
		case !m.showHelp && m.matches("page_jukebox", msg):
			return m.switchPage(jukeboxPage)
		case !m.showHelp && m.matches("compose_post", msg):
			return m.openComposer()
		case !m.showHelp && m.matches("switch_theme", msg):
			return m.switchTheme()
		case !m.showHelp && m.page == feedPage && m.matches("select_next", msg):
			if len(m.posts) > 0 {
				m.selectedPost = min(m.selectedPost+1, len(m.posts)-1)
				m.renderFeed()
			}
			return m, nil
		case !m.showHelp && m.page == feedPage && m.matches("select_previous", msg):
			if len(m.posts) > 0 {
				m.selectedPost = max(m.selectedPost-1, 0)
				m.renderFeed()
			}
			return m, nil
		case !m.showHelp && m.page == feedPage && m.matches("open_post", msg) && len(m.posts) > 0:
			m.loading, m.err, m.notice = true, nil, ""
			return m, m.loadPostDetail(m.posts[m.selectedPost].PostID)
		case !m.showHelp && m.page == feedPage && m.matches("toggle_bookmark", msg) && len(m.posts) > 0:
			m.loading, m.err, m.notice = true, nil, ""
			return m, m.bookmarkPost(m.posts[m.selectedPost].PostID)
		case !m.showHelp && m.page == jukeboxPage && m.matches("jukebox_select_next", msg):
			return m.jukeboxMove(1)
		case !m.showHelp && m.page == jukeboxPage && m.matches("jukebox_select_previous", msg):
			return m.jukeboxMove(-1)
		case !m.showHelp && m.page == jukeboxPage && m.matches("jukebox_play", msg):
			return m.jukeboxPlay()
		case !m.showHelp && m.page == jukeboxPage && m.matches("jukebox_pause", msg):
			return m.jukeboxPause()
		case !m.showHelp && m.page == jukeboxPage && m.matches("jukebox_next", msg):
			return m.jukeboxNext()
		case !m.showHelp && m.page == jukeboxPage && m.matches("jukebox_previous", msg):
			return m.jukeboxPrevious()
		case !m.showHelp && m.page == jukeboxPage && m.matches("jukebox_stop", msg):
			return m.jukeboxStop()
		case !m.showHelp && m.page == jukeboxPage && m.matches("jukebox_page_next", msg) && !m.loading:
			return m.jukeboxNextPage()
		case !m.showHelp && m.page == jukeboxPage && m.matches("jukebox_page_previous", msg):
			return m.jukeboxPreviousPage()
		case !m.showHelp && m.page == jukeboxPage && m.matches("back", msg):
			m.page = feedPage
			m.err = nil
			m.viewport.GotoTop()
			m.renderFeed()
			return m, nil
		case !m.showHelp && m.matches("refresh", msg) && !m.loading:
			m.loading, m.err = true, nil
			return m, m.loadCurrentPage()
		case !m.showHelp && m.page == feedPage && m.matches("next_page", msg) && !m.loading && m.cursor != "":
			m.loading, m.err = true, nil
			return m, m.loadFeed(m.cursor, false)
		case !m.showHelp && m.matches("scroll_up", msg):
			m.viewport.ScrollUp(1)
			return m, nil
		case !m.showHelp && m.matches("scroll_down", msg):
			m.viewport.ScrollDown(1)
			return m, nil
		case !m.showHelp && m.matches("page_up", msg):
			m.viewport.ScrollUp(max(m.viewport.Height()-1, 1))
			return m, nil
		case !m.showHelp && m.matches("page_down", msg):
			m.viewport.ScrollDown(max(m.viewport.Height()-1, 1))
			return m, nil
		case !m.showHelp && m.matches("top", msg):
			m.viewport.GotoTop()
			return m, nil
		case !m.showHelp && m.matches("bottom", msg):
			m.viewport.GotoBottom()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// header renders the wordmark and status line centered at the top of the
// screen. The glyphs are spread out so the brand line reads as a display
// headline rather than a plain label.
func (m Model) header() string {
	spread := strings.Join(strings.Split(wordmark, ""), " ")
	title := titleStyle.Width(m.width).Align(lipgloss.Center).Render(spread)
	meta := metaStyle.Width(m.width).Align(lipgloss.Center).Render("@" + m.client.Username + "  ·  " + string(m.page) + "  ·  " + m.theme)
	return lipgloss.JoinVertical(lipgloss.Left, title, meta)
}

func (m Model) View() tea.View {
	if m.width == 0 {
		view := tea.NewView("Loading Cyberspace…")
		view.AltScreen = true
		return view
	}

	header := m.header()
	if m.composing {
		view := tea.NewView(m.composerView(header))
		view.AltScreen = true
		return view
	}
	if m.viewingPost {
		view := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, header, m.viewport.View(), helpStyle.Render(m.postDetailHelp())))
		view.AltScreen = true
		return view
	}
	footer := helpStyle.Render(m.helpLine())
	switch {
	case m.loading:
		footer = helpStyle.Render("Loading feed…")
	case m.err != nil:
		footer = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6b6b")).Render("Could not load feed: " + m.err.Error())
	case m.notice != "":
		footer = titleStyle.Render(m.notice)
	case m.page == jukeboxPage:
		footer = helpStyle.Render(m.jukeboxStatusLine())
	}

	content := lipgloss.JoinVertical(lipgloss.Left, header, m.viewport.View(), footer)
	if m.showHelp {
		content = lipgloss.JoinVertical(lipgloss.Left, header, m.helpView())
	}
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func (m Model) openComposer() (tea.Model, tea.Cmd) {
	return m.openComposerForReply(false)
}

func (m Model) openReplyComposer() (tea.Model, tea.Cmd) {
	return m.openComposerForReply(true)
}

func (m Model) openComposerForReply(reply bool) (tea.Model, tea.Cmd) {
	composer := textarea.New()
	composer.Placeholder = "Write a post…"
	composer.ShowLineNumbers = false
	composer.SetWidth(max(m.width-6, 30))
	composer.SetHeight(max(m.height-8, 6))
	m.composer, m.composing, m.confirmPost, m.replying, m.err = composer, true, false, reply, nil
	m.replyParentID = ""
	m.replyParentAuthor = ""
	if reply {
		composer.Placeholder = "Write a reply…"
		m.replyParentAuthor = m.activePost.AuthorUsername
		if m.replyIdx >= 0 && m.replyIdx < len(m.replies) {
			m.replyParentID = m.replies[m.replyIdx].ReplyID
			m.replyParentAuthor = m.replies[m.replyIdx].AuthorUsername
		}
	}
	return m, m.composer.Focus()
}

func (m Model) switchTheme() (tea.Model, tea.Cmd) {
	current := 0
	for i, candidate := range themes {
		if candidate.name == m.theme {
			current = i
			break
		}
	}
	m.theme = themes[(current+1)%len(themes)].name
	applyTheme(m.theme)
	m.client.Config.Settings.Theme = m.theme
	if err := m.client.SaveConfig(m.client.Config); err != nil {
		m.err = err
	}
	m.renderCurrentPage()
	return m, nil
}

func (m Model) updateComposer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.posting {
		return m, nil
	}
	if m.confirmPost {
		switch {
		case m.matches("cancel_compose", msg):
			m.confirmPost = false
			return m, nil
		case m.matches("confirm_post", msg):
			m.posting = true
			if m.replying {
				return m, m.createReply(m.composer.Value())
			}
			return m, m.createPost(m.composer.Value())
		}
		return m, nil
	}
	if m.matches("cancel_compose", msg) {
		m.composing = false
		return m, nil
	}
	if m.matches("submit_post", msg) {
		if strings.TrimSpace(m.composer.Value()) == "" {
			m.err = fmt.Errorf("a post cannot be empty")
			return m, nil
		}
		m.confirmPost = true
		m.composer.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.composer, cmd = m.composer.Update(msg)
	return m, cmd
}

func (m Model) updatePostDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case m.matches("back", msg):
		m.viewingPost = false
		m.viewport.GotoTop()
		m.renderFeed()
		return m, nil
	case m.matches("select_next", msg):
		if len(m.replies) > 0 {
			m.replyIdx = min(m.replyIdx+1, len(m.replies)-1)
			m.renderPostDetail()
		}
		return m, nil
	case m.matches("select_previous", msg):
		m.replyIdx = max(m.replyIdx-1, -1)
		m.renderPostDetail()
		return m, nil
	case m.matches("open_post", msg):
		return m.openReplyComposer()
	case m.matches("toggle_bookmark", msg) && !m.loading:
		m.loading, m.err, m.notice = true, nil, ""
		return m, m.bookmarkPost(m.activePost.PostID)
	case m.matches("reply_to_post", msg):
		return m.openReplyComposer()
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) loadPostDetail(postID string) tea.Cmd {
	return func() tea.Msg {
		post, err := m.client.GetPostById(postID)
		if err != nil {
			return postDetailLoadedMsg{err: err}
		}
		replies, _, err := m.client.GetReplies(postID, 50, "")
		return postDetailLoadedMsg{post: post, replies: replies, err: err}
	}
}

func (m Model) bookmarkPost(postID string) tea.Cmd {
	return func() tea.Msg { return bookmarkCreatedMsg{err: m.client.CreateBookmark(postID, "post")} }
}

func (m Model) jukeboxPageSize() int {
	if m.height > 12 {
		return max(m.height-8, 4)
	}
	return 10
}

func (m Model) jukeboxPageStart() int {
	return (m.jukeboxPage - 1) * m.jukeboxPageSize()
}

func (m Model) jukeboxPageEnd() int {
	return min(m.jukeboxPageStart()+m.jukeboxPageSize(), len(m.tracks))
}

func (m Model) jukeboxLastPage() int {
	if len(m.tracks) == 0 {
		return 1
	}
	return (len(m.tracks)-1)/m.jukeboxPageSize() + 1
}

func (m Model) jukeboxMove(delta int) (tea.Model, tea.Cmd) {
	if len(m.tracks) == 0 {
		return m, nil
	}
	start, end := m.jukeboxPageStart(), m.jukeboxPageEnd()
	m.jukeboxIdx = min(max(m.jukeboxIdx+delta, start), end-1)
	m.playerErr = nil
	m.renderJukebox()
	m.viewport.EnsureVisible(m.jukeboxIdx, 0, m.width)
	return m, nil
}

// jukeboxNextPage advances to the next page of the catalogue, fetching a
// deeper slice of the feed first if the current page is the last one loaded.
func (m Model) jukeboxNextPage() (tea.Model, tea.Cmd) {
	if len(m.tracks) == 0 {
		return m, nil
	}
	if m.jukeboxPage < m.jukeboxLastPage() {
		m.jukeboxPage++
		m.jukeboxIdx = m.jukeboxPageStart()
		m.playerErr = nil
		m.renderJukebox()
		m.viewport.GotoTop()
		return m, nil
	}
	if m.tracksCursor == "" {
		return m, nil
	}
	m.loading, m.err, m.jukeboxAdvancePage = true, nil, true
	return m, m.loadJukeboxTracks(false)
}

// jukeboxPreviousPage steps back a page within the already-loaded catalogue.
func (m Model) jukeboxPreviousPage() (tea.Model, tea.Cmd) {
	if m.jukeboxPage <= 1 {
		return m, nil
	}
	m.jukeboxPage--
	m.jukeboxIdx = m.jukeboxPageStart()
	m.playerErr = nil
	m.renderJukebox()
	m.viewport.GotoTop()
	return m, nil
}

func (m Model) jukeboxPlay() (tea.Model, tea.Cmd) {
	if len(m.tracks) == 0 {
		return m, nil
	}
	m.playerErr = nil
	m.viewport.EnsureVisible(m.jukeboxIdx, 0, m.width)
	return m, m.playerCmd("play", m.jukeboxIdx)
}

func (m Model) jukeboxPause() (tea.Model, tea.Cmd) {
	if len(m.tracks) == 0 || m.nowPlaying < 0 {
		m.playerErr = fmt.Errorf("no track is playing")
		return m, nil
	}
	return m, m.playerCmd("pause", -1)
}

func (m Model) jukeboxNext() (tea.Model, tea.Cmd) {
	if len(m.tracks) == 0 {
		return m, nil
	}
	m.playerErr = nil
	next := m.nowPlaying + 1
	if next < 0 || next >= len(m.tracks) {
		next = 0
	}
	m.jukeboxIdx = next
	m.renderJukebox()
	m.viewport.EnsureVisible(m.jukeboxIdx, 0, m.width)
	return m, m.playerCmd("play", next)
}

func (m Model) jukeboxPrevious() (tea.Model, tea.Cmd) {
	if len(m.tracks) == 0 {
		return m, nil
	}
	m.playerErr = nil
	previous := m.nowPlaying - 1
	if previous < 0 {
		previous = len(m.tracks) - 1
	}
	m.jukeboxIdx = previous
	m.renderJukebox()
	m.viewport.EnsureVisible(m.jukeboxIdx, 0, m.width)
	return m, m.playerCmd("play", previous)
}

func (m Model) jukeboxStop() (tea.Model, tea.Cmd) {
	return m, m.playerCmd("stop", -1)
}

func (m Model) playerCmd(action string, index int) tea.Cmd {
	return func() tea.Msg {
		msg := playerStatusMsg{action: action, index: index, paused: m.paused}
		var err error
		switch action {
		case "play":
			if index < 0 || index >= len(m.tracks) {
				msg.err = fmt.Errorf("no track selected")
				return msg
			}
			msg.paused = false
			err = m.player.Play(m.tracks[index])
		case "pause":
			err = m.player.TogglePause()
			if err == nil {
				msg.paused = m.player.paused
			}
		case "stop":
			m.player.Stop()
		}
		msg.err = err
		return msg
	}
}

func (m *Model) renderJukebox() {
	if len(m.tracks) == 0 {
		m.viewport.SetContent(metaStyle.Render("No audio tracks in the feed yet. Press " + m.keyNames("refresh") + " to refresh and index new posts."))
		return
	}
	start, end := m.jukeboxPageStart(), m.jukeboxPageEnd()
	rows := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		track := m.tracks[i]
		marker := "  "
		if i == m.jukeboxIdx {
			marker = "▶ "
		}
		line := fmt.Sprintf("%s%d. %s", marker, i+1, track.Label())
		if track.Genre != "" {
			line += "  ·  " + track.Genre
		}
		rows = append(rows, postStyle.Width(max(m.width-6, 24)).Render(line))
	}
	if m.tracksCursor != "" {
		rows = append(rows, metaStyle.Render("More tracks further back in the feed: press "+m.keyNames("jukebox_page_next")+" to load the next page."))
	}
	m.viewport.SetContent(strings.Join(rows, "\n"))
}

func (m Model) jukeboxStatusLine() string {
	if m.playerErr != nil {
		return m.playerErr.Error()
	}
	if m.nowPlaying >= 0 && m.nowPlaying < len(m.tracks) {
		state := "playing"
		if m.paused {
			state = "paused"
		}
		return fmt.Sprintf("♪ %s  ·  %s  ·  %s play  %s pause  %s next  %s previous  %s stop",
			m.tracks[m.nowPlaying].Label(), state,
			m.keyNames("jukebox_play"), m.keyNames("jukebox_pause"),
			m.keyNames("jukebox_next"), m.keyNames("jukebox_previous"), m.keyNames("jukebox_stop"))
	}
	nav := fmt.Sprintf("%s prev page  %s next page", m.keyNames("jukebox_page_previous"), m.keyNames("jukebox_page_next"))
	return fmt.Sprintf("%d tracks  ·  page %d/%d  ·  %s play  %s pause  %s stop  ·  %s  ·  %s back  %s refresh",
		len(m.tracks), m.jukeboxPage, m.jukeboxLastPage(),
		m.keyNames("jukebox_play"), m.keyNames("jukebox_pause"), m.keyNames("jukebox_stop"),
		nav, m.keyNames("back"), m.keyNames("refresh"))
}

func (m *Model) renderPostDetail() {
	rendered := renderPost(m.activePost, m.width)
	if m.replyIdx == -1 {
		rendered = titleStyle.Render("▶ selected") + "\n" + rendered
	}
	parts := []string{rendered}
	if len(m.replies) == 0 {
		parts = append(parts, metaStyle.Render("No replies yet."))
	}
	for i, reply := range m.replies {
		rendered := renderReply(reply, m.width)
		if i == m.replyIdx {
			rendered = titleStyle.Render("▶ selected") + "\n" + rendered
		}
		parts = append(parts, rendered)
	}
	m.viewport.SetContent(strings.Join(parts, "\n\n"))
}

func (m Model) postDetailHelp() string {
	if m.loading {
		return "Loading post…"
	}
	if m.err != nil {
		return m.err.Error()
	}
	if m.notice != "" {
		return m.notice
	}
	return fmt.Sprintf("%s back  ·  %s select reply  ·  %s reply to selected  ·  %s bookmark  ·  ↑/↓ scroll",
		m.keyNames("back"), m.keyNames("select_next"), m.keyNames("open_post"), m.keyNames("toggle_bookmark"))
}

func (m Model) createPost(content string) tea.Cmd {
	return func() tea.Msg {
		post, err := m.client.CreatePost(client.CreatePostInput{Content: content})
		return postCreatedMsg{post: post, err: err}
	}
}

func (m Model) createReply(content string) tea.Cmd {
	return func() tea.Msg {
		reply, err := m.client.CreateReply(client.CreateReplyInput{
			PostID:        m.activePost.PostID,
			Content:       content,
			ParentReplyID: m.replyParentID,
		})
		return replyCreatedMsg{reply: reply, err: err}
	}
}

func (m Model) composerView(header string) string {
	footer := metaStyle.Render(m.keyNames("submit_post") + " submit  ·  " + m.keyNames("cancel_compose") + " cancel")
	if m.confirmPost {
		footer = titleStyle.Render("Post this? ") + metaStyle.Render(m.keyNames("confirm_post")+" confirm  ·  "+m.keyNames("cancel_compose")+" return to editing")
	}
	if m.posting {
		footer = metaStyle.Render("Posting…")
	}
	if m.err != nil {
		footer = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6b6b")).Render(m.err.Error())
	}
	title := "Compose post"
	if m.replying {
		title = "Write reply"
		if m.replyParentAuthor != "" {
			title = "Write reply to @" + m.replyParentAuthor
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, titleStyle.Render(title), m.composer.View(), footer)
}

func (m Model) matches(action string, msg tea.KeyMsg) bool {
	return slices.Contains(m.keys[action], msg.String())
}

func (m Model) helpLine() string {
	return fmt.Sprintf("%s help  ·  %s refresh  ·  %s older posts  ·  %s quit",
		m.keyNames("help"), m.keyNames("refresh"), m.keyNames("next_page"), m.keyNames("quit"))
}

func (m Model) helpView() string {
	rows := []string{
		titleStyle.Render("Keyboard shortcuts"),
		"",
		m.helpRow("page_feed", "feed"),
		m.helpRow("page_bookmarks", "bookmarks"),
		m.helpRow("page_notifications", "notifications"),
		m.helpRow("page_journal", "journal"),
		m.helpRow("page_profile", "profile"),
		m.helpRow("page_mail", "C-Mail"),
		m.helpRow("page_jukebox", "jukebox"),
		m.helpRow("compose_post", "compose post"),
		m.helpRow("submit_post", "review post"),
		m.helpRow("switch_theme", "cycle theme"),
		m.helpRow("select_next", "select next feed post / reply"),
		m.helpRow("select_previous", "select previous feed post / reply"),
		m.helpRow("open_post", "open post / reply to selected"),
		m.helpRow("toggle_bookmark", "bookmark selected post"),
		m.helpRow("back", "back to feed (post detail, jukebox)"),
		m.helpRow("jukebox_select_next", "select next track"),
		m.helpRow("jukebox_select_previous", "select previous track"),
		m.helpRow("jukebox_play", "play selected track"),
		m.helpRow("jukebox_pause", "pause/resume"),
		m.helpRow("jukebox_next", "next track"),
		m.helpRow("jukebox_previous", "previous track"),
		m.helpRow("jukebox_stop", "stop playback"),
		m.helpRow("jukebox_page_next", "next jukebox page"),
		m.helpRow("jukebox_page_previous", "previous jukebox page"),
		m.helpRow("scroll_up", "scroll up"),
		m.helpRow("scroll_down", "scroll down"),
		m.helpRow("page_up", "previous page"),
		m.helpRow("page_down", "next page"),
		m.helpRow("top", "go to top"),
		m.helpRow("bottom", "go to bottom"),
		m.helpRow("refresh", "refresh feed"),
		m.helpRow("next_page", "load older posts / more tracks"),
		m.helpRow("quit", "quit"),
		"",
		metaStyle.Render("Edit settings.keybindings in config.json to remap any action."),
		metaStyle.Render("Press " + m.keyNames("close_help") + " to return."),
	}
	return postStyle.Width(max(m.width-6, 36)).Render(strings.Join(rows, "\n"))
}

func (m Model) switchPage(page pageID) (tea.Model, tea.Cmd) {
	if m.page == page {
		return m, nil
	}
	m.page, m.loading, m.err = page, true, nil
	m.viewport.GotoTop()
	return m, m.loadCurrentPage()
}

func (m Model) loadCurrentPage() tea.Cmd {
	switch m.page {
	case feedPage:
		return m.loadFeed("", true)
	case notificationsPage:
		return func() tea.Msg {
			items, _, err := m.client.GetNotifications(25, "")
			return notificationsLoadedMsg{items: items, err: err}
		}
	case bookmarksPage:
		return func() tea.Msg {
			items, _, err := m.client.GetBookmarks(25, "")
			return bookmarksLoadedMsg{items: items, err: err}
		}
	case journalPage:
		return func() tea.Msg {
			items, _, err := m.client.GetNotes(25, "")
			return notesLoadedMsg{items: items, err: err}
		}
	case profilePage:
		return func() tea.Msg {
			user, err := m.client.GetMyUserProfile()
			return profileLoadedMsg{user: user, err: err}
		}
	case jukeboxPage:
		return m.loadJukeboxTracks(true)
	default:
		return func() tea.Msg { return pageLoadedPlaceholderMsg{} }
	}
}

type pageLoadedPlaceholderMsg struct{}

func (m *Model) renderCurrentPage() {
	switch m.page {
	case feedPage:
		m.renderFeed()
	case notificationsPage:
		m.viewport.SetContent(renderNotifications(m.notifications, m.width))
	case bookmarksPage:
		m.viewport.SetContent(renderBookmarks(m.bookmarks, m.width))
	case journalPage:
		m.viewport.SetContent(renderNotes(m.notes, m.width))
	case profilePage:
		m.viewport.SetContent(renderProfile(m.profile, m.width))
	case mailPage:
		m.viewport.SetContent(renderUnavailable("C-Mail", "The modern site provides C-Mail at /messages. Its Firebase conversation service is the next integration phase."))
	case jukeboxPage:
		m.renderJukebox()
	}
}

func renderNotifications(items []client.Notification, width int) string {
	if len(items) == 0 {
		return metaStyle.Render("No notifications to show.")
	}
	rows := make([]string, 0, len(items))
	for _, item := range items {
		state := "read"
		if !item.Read {
			state = "new"
		}
		rows = append(rows, postStyle.Width(max(width-6, 24)).Render(fmt.Sprintf("%s  ·  %s  ·  %s", titleStyle.Render("@"+item.ActorUsername), item.Type, state)+"\n"+metaStyle.Render(relativeTime(item.CreatedAt))))
	}
	return strings.Join(rows, "\n\n")
}

func renderBookmarks(items []client.Bookmark, width int) string {
	if len(items) == 0 {
		return metaStyle.Render("No bookmarks to show.")
	}
	rows := make([]string, 0, len(items))
	for _, item := range items {
		target := item.PostID
		if item.Type == "reply" {
			target = item.ReplyID
		}
		rows = append(rows, postStyle.Width(max(width-6, 24)).Render(titleStyle.Render("Saved "+item.Type)+"\n"+metaStyle.Render("Target: "+target+"  ·  "+relativeTime(item.CreatedAt))))
	}
	return strings.Join(rows, "\n\n")
}

func renderNotes(items []client.Note, width int) string {
	if len(items) == 0 {
		return metaStyle.Render("Your journal is empty.")
	}
	rows := make([]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, postStyle.Width(max(width-6, 24)).Render(metaStyle.Render(relativeTime(item.CreatedAt))+"\n"+wrap(item.Content, max(width-10, 24))))
	}
	return strings.Join(rows, "\n\n")
}

func renderProfile(user client.User, width int) string {
	if user.Username == "" {
		return metaStyle.Render("Profile unavailable.")
	}
	rows := []string{titleStyle.Render("@" + user.Username), metaStyle.Render(user.Bio)}
	rows = append(rows, fmt.Sprintf("%d posts  ·  %d followers  ·  %d following", user.PostsCount, user.FollowersCount, user.FollowingCount))
	if user.WebsiteURL != "" {
		rows = append(rows, metaStyle.Render(user.WebsiteURL))
	}
	return postStyle.Width(max(width-6, 24)).Render(strings.Join(rows, "\n"))
}

func renderUnavailable(name, description string) string {
	return postStyle.Render(titleStyle.Render(name) + "\n" + metaStyle.Render(description))
}

func (m Model) helpRow(action, description string) string {
	return fmt.Sprintf("%-18s %s", m.keyNames(action), description)
}

func (m Model) keyNames(action string) string {
	keys := m.keys[action]
	if len(keys) == 0 {
		return "disabled"
	}
	return strings.Join(keys, "/")
}

func (m Model) loadFeed(cursor string, reset bool) tea.Cmd {
	return func() tea.Msg {
		posts, nextCursor, err := m.client.GetPosts(10, cursor)
		return feedLoadedMsg{posts: posts, cursor: nextCursor, reset: reset, err: err}
	}
}

// loadJukeboxTracks scans a window of the feed so the jukebox finds audio
// posts that sit further back than the first feed page. A reset scan starts
// from the beginning and replaces the catalogue; a non-reset scan continues
// from where the last one stopped and appends new tracks.
func (m Model) loadJukeboxTracks(reset bool) tea.Cmd {
	return func() tea.Msg {
		cursor := ""
		if !reset {
			cursor = m.tracksCursor
		}
		posts, next, err := m.client.GetPostsForTracks(200, cursor)
		return jukeboxLoadedMsg{tracks: audioTracksFromPosts(posts), cursor: next, reset: reset, err: err}
	}
}

func (m *Model) resizeViewport() {
	if m.width == 0 || m.height == 0 {
		return
	}
	m.viewport.SetWidth(m.width)
	m.viewport.SetHeight(max(m.height-3, 1))
}

func (m *Model) renderFeed() {
	if m.width == 0 {
		return
	}
	if len(m.posts) == 0 && !m.loading {
		m.viewport.SetContent(metaStyle.Render("No posts to show."))
		return
	}

	posts := make([]string, 0, len(m.posts))
	for index, post := range m.posts {
		if post.IsNSFW {
			continue
		}
		rendered := renderPost(post, m.width)
		if index == m.selectedPost {
			rendered = titleStyle.Render("▶ selected") + "\n" + rendered
		}
		posts = append(posts, rendered)
	}
	m.viewport.SetContent(strings.Join(posts, "\n\n"))
}

func renderPost(post client.Post, width int) string {
	contentWidth := max(width-6, 24)
	content := strings.TrimSpace(post.Content)
	if len([]rune(content)) > 1_000 {
		content = string([]rune(content)[:1_000]) + "…"
	}

	meta := relativeTime(post.CreatedAt)
	if post.BookmarksCount > 0 {
		meta += fmt.Sprintf("  ·  %d saves", post.BookmarksCount)
	}
	if post.RepliesCount > 0 {
		meta += fmt.Sprintf("  ·  %d replies", post.RepliesCount)
	}

	parts := []string{titleStyle.Render("@" + post.AuthorUsername), metaStyle.Render(meta), wrap(content, contentWidth)}
	if len(post.Topics) > 0 {
		parts = append(parts, metaStyle.Render(strings.Join(post.Topics, "  ")))
	}
	return postStyle.Width(contentWidth).Render(strings.Join(parts, "\n"))
}

func renderReply(reply client.Reply, width int) string {
	contentWidth := max(width-8, 24)
	meta := relativeTime(reply.CreatedAt)
	if reply.SavesCount > 0 {
		meta += fmt.Sprintf("  ·  %d saves", reply.SavesCount)
	}
	if reply.ParentReplyAuthor != "" {
		meta += "  ·  replying to @" + reply.ParentReplyAuthor
	}
	parts := []string{titleStyle.Render("↳ @" + reply.AuthorUsername), metaStyle.Render(meta), wrap(reply.Content, contentWidth)}
	return postStyle.Width(contentWidth).Render(strings.Join(parts, "\n"))
}

func relativeTime(at time.Time) string {
	if at.IsZero() {
		return "now"
	}
	d := time.Since(at)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func wrap(text string, width int) string {
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		line := words[0]
		for _, word := range words[1:] {
			if len([]rune(line))+len([]rune(word))+1 > width {
				lines = append(lines, line)
				line = word
				continue
			}
			line += " " + word
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
