package tui

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	client "github.com/johannesalke/cyberspacecli/internal/cyberspaceClient"
)

// Track is a playable audio attachment found in the Cyberspace feed.
type Track struct {
	Source string
	Title  string
	Artist string
	Genre  string
	PostID string
}

func (t Track) Label() string {
	if t.Artist != "" && t.Title != "" {
		return t.Artist + " — " + t.Title
	}
	if t.Title != "" {
		return t.Title
	}
	return t.Source
}

// audioPlayer controls a background mpv process through its local IPC socket.
// The TUI remains responsible for all visible player state and controls.
type audioPlayer struct {
	mu      sync.Mutex
	command *exec.Cmd
	socket  string
	track   Track
	paused  bool
}

func newAudioPlayer() *audioPlayer {
	return &audioPlayer{socket: filepath.Join(os.TempDir(), fmt.Sprintf("cyberspace-mpv-%d.sock", os.Getpid()))}
}

func (p *audioPlayer) Play(track Track) error {
	if track.Source == "" {
		return fmt.Errorf("this track has no playable source")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopLocked()
	p.command = exec.Command("mpv", "--no-video", "--force-window=no", "--really-quiet", "--input-ipc-server="+p.socket, "--", track.Source)
	if err := p.command.Start(); err != nil {
		return fmt.Errorf("start mpv: %w", err)
	}
	p.track, p.paused = track, false
	go func(command *exec.Cmd) {
		_ = command.Wait()
	}(p.command)
	return nil
}

func (p *audioPlayer) TogglePause() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.command == nil || p.command.Process == nil {
		return fmt.Errorf("no track is playing")
	}
	// mpv may take a moment to create its socket after process start.
	var err error
	for range 5 {
		err = p.send(`{"command":["cycle","pause"]}`)
		if err == nil {
			p.paused = !p.paused
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("pause track: %w", err)
}

func (p *audioPlayer) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopLocked()
}

func (p *audioPlayer) stopLocked() {
	if p.command == nil {
		return
	}
	_ = p.send(`{"command":["quit"]}`)
	if p.command.Process != nil {
		_ = p.command.Process.Kill()
	}
	p.command = nil
	p.paused = false
	_ = os.Remove(p.socket)
}

func (p *audioPlayer) send(message string) error {
	connection, err := net.DialTimeout("unix", p.socket, 200*time.Millisecond)
	if err != nil {
		return err
	}
	defer connection.Close()
	_, err = connection.Write([]byte(message + "\n"))
	return err
}

// audioTracksFromPosts extracts every audio attachment from a page of posts,
// skipping any post whose attachment metadata cannot be decoded.
func audioTracksFromPosts(posts []client.Post) []Track {
	var tracks []Track
	for _, post := range posts {
		audio, err := post.AudioAttachments()
		if err != nil {
			continue
		}
		for _, a := range audio {
			tracks = append(tracks, Track{
				Source: a.Src,
				Title:  a.Title,
				Artist: a.Artist,
				Genre:  a.Genre,
				PostID: post.PostID,
			})
		}
	}
	return tracks
}

// mergeTracks appends incoming tracks to an existing catalogue, de-duplicating
// by source URL so refreshing the feed never doubles a song.
func mergeTracks(existing, incoming []Track) []Track {
	seen := make(map[string]bool, len(existing)+len(incoming))
	out := append([]Track(nil), existing...)
	for _, track := range out {
		seen[track.Source] = true
	}
	for _, track := range incoming {
		if track.Source != "" && !seen[track.Source] {
			seen[track.Source] = true
			out = append(out, track)
		}
	}
	return out
}
