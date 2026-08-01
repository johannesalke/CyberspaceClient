```
 ██████╗██╗   ██╗██████╗ ███████╗██████╗ ███████╗██████╗  █████╗  ██████╗███████╗
██╔════╝╚██╗ ██╔╝██╔══██╗██╔════╝██╔══██╗██╔════╝██╔══██╗██╔══██╗██╔════╝██╔════╝
██║      ╚████╔╝ ██████╔╝█████╗  ██████╔╝███████╗██████╔╝███████║██║     █████╗
██║       ╚██╔╝  ██╔══██╗██╔══╝  ██╔══██╗╚════██║██╔═══╝ ██╔══██║██║     ██╔══╝
╚██████╗   ██║   ██████╔╝███████╗██║  ██║███████║██║     ██║  ██║╚██████╗███████╗
 ╚═════╝   ╚═╝   ╚═════╝ ╚══════╝╚═╝  ╚═╝╚══════╝╚═╝     ╚═╝  ╚═╝ ╚═════╝╚══════╝
```

# Cyberspace CLI Client

This is a Commandline Client for the social network platform [Cyberspace](https://cyberspace.online/). It is currently in an alpha state.

At present, this client has solid basic functions and presentation, but lacks the full functionality of the website version.
You can browse your feed and notifications, write posts and replies as well as write, edit and publish notes.

All you need is a Cyberspace account: log in with your email and password on first launch.


![Graphical Showcase](https://i.postimg.cc/kgwgddBQ/github-showcase-1.png)


## Quick Start

### Install the TUI

The recommended client is a full-screen, keyboard-navigable terminal interface. There are three ways to install it:

**From a GitHub release (no clone, no Go):**

```sh
curl -fsSL https://raw.githubusercontent.com/johannesalke/cyberspacecli/main/install.sh | sh
```

This downloads the prebuilt binary for your OS/architecture into `~/.local/bin`. Use `install.sh v1.2.3` to pin a tag, or `--source` to build from source instead.

**With Go, without cloning:**

```sh
go install github.com/johannesalke/cyberspacecli/cmd/cyberspace@latest
```

**From a clone:**

```sh
make build     # builds ./bin/cyberspace
make install   # installs to ~/.local/bin
```

Then launch it from any directory:

```sh
cyberspace
```

Use `↑`/`↓` to scroll, `r` to refresh, `n` to load older posts, and `q` to exit. The command is installed to `~/.local/bin`; that directory is on `PATH` in most setups. To remove the installed command later, run `make uninstall`.

### Building a release

`make release` cross-compiles static binaries for Linux (amd64/arm64), macOS (amd64/arm64), and Windows (amd64) into `dist/` as tarballs. `make` derives the version stamp from the nearest `git` tag. Publishing a `v*` tag triggers the GitHub Actions workflow (`.github/workflows/release.yml`), which builds the tarballs and attaches them to a GitHub Release; `install.sh` then installs them without a clone.

### Keyboard navigation and remapping

The TUI is fully usable from the keyboard. Press `?` at any time to see the active bindings. By default it supports arrow keys and `j`/`k` to scroll, Page Up/Page Down (or `ctrl+u`/`ctrl+d`) for page movement, `g`/Home for the top, and `G`/End for the bottom.

Bindings live in the `settings.keybindings` section of the client config file (`~/.config/cyberspace_client/config.json` on Linux unless `XDG_CONFIG_HOME` is set). Every action accepts a list of terminal key names; replacing a list remaps that action, and an empty list disables it. For example:

```json
{
  "settings": {
    "keybindings": {
      "scroll_up": ["w"],
      "scroll_down": ["s"],
      "refresh": ["ctrl+r"],
      "quit": ["ctrl+q"],
      "next_page": []
    }
  }
}
```

Supported actions are `quit`, `help`, `close_help`, `refresh`, `next_page`, `scroll_up`, `scroll_down`, `page_up`, `page_down`, `top`, `bottom`, `page_feed`, `page_bookmarks`, `page_notifications`, `page_journal`, `page_profile`, `page_mail`, `page_jukebox`, `compose_post`, `submit_post`, `confirm_post`, `cancel_compose`, `switch_theme`, `select_next`, `select_previous`, `focus_notifications`, `open_post`, `back`, `toggle_bookmark`, `reply_to_post`, `jukebox_select_next`, `jukebox_select_previous`, `jukebox_play`, `jukebox_pause`, `jukebox_next`, `jukebox_previous`, `jukebox_stop`, `jukebox_page_next`, and `jukebox_page_previous`. Missing actions retain their defaults.

### Using the TUI

The interface is organized into seven pages, one per number key:

- `1` Feed · `2` Bookmarks · `3` Notifications · `4` Journal · `5` Profile · `6` C-Mail · `7` Jukebox

**Posting** — press `c` to open the composer, type your message, `ctrl+s` to review it, then `enter`/`y` to confirm or `esc` to cancel.

**Replying** — `enter` opens the selected post with its replies. `tab`/`shift+tab` (or `j`/`k`) moves the selection across the replies, and `enter` or `r` opens the composer addressed to the reply (or post) you selected. `b` bookmarks the open post.

**Feed sidebar** — on wide terminals the feed page carries a right-hand panel. Its top half lists recent notifications, and its bottom half shows the jukebox's now-playing track. Press `N` to move keyboard focus into the notifications list: `tab`/`shift+tab` (or the movement keys) cycle the selection, `enter` opens the notification's post or reply (marking it read), and `esc` returns focus to the feed. Tabbing through feed posts keeps the selection in view by scrolling the feed automatically, so a selected post is never stranded off-screen.

**Player anywhere** — while a track is playing, the non-conflicting player keys work from the feed page too: `p` pauses, `x` stops, `←`/`→` skip tracks. `→` loads older posts instead while there are still pages of the feed left to read.

**Jukebox** — the catalogue is built by scanning the feed for audio attachments. `enter`/`space` plays the selected track, `p` pauses, `x` stops, `←`/`→` skip tracks, and `n`/`pgdown` + `pgup` page through the catalogue. Paging past the last loaded page scans deeper into the feed for more tracks. Playback goes through `mpv` (with `yt-dlp` resolving stream URLs), so both need to be installed for the jukebox to play.

**Themes** — `t` cycles through the color themes; your choice is saved to the config file.

**Version** — `cyberspace --version` prints the build version.

### Legacy command-line client

This assumes you already have a Go itself installed. If you do not, refer to this page [this page](https://golang.org/doc/install) before returning.

So long as you have the client installed, you can simply clone the Git (`git clone github.com/johannesalke/cyberspacecli`) repo onto your machine (or download it via github), then while inside the project directory execute the following commands:

```go
go build -o cyberspacecli .
./cyberspacecli
```
The first comiles the program, the second executes it. (To exit, use the command 'exit' or press the keys ctrl+C.)

If you like what you see and want to have the tool available anywhere on you machine (as opposed to just inside this folder) you can also install it as a command line tool. You can then use it as easily as you would git or vim. For that, see the following instructions:

```
go install github.com/johannesalke/cyberspacecli
```

## Usage

Client commands consist of a verb and a noun.


- `view feed (optional_arg)`: Load 10 posts from the feed, starting at the newest. Every time the command is used, 10 more are loaded starting from where the previous iteration stopped. In the feed, posts are truncated at 1000 characters. To see the whole post, use the 'view post' command.
Use the optional argument 'new' to load posts made since you started the client without losing the marker of the basic command. Use 'reset' to start over entirely. (Sometimes it may show less than 10 because NSFW posts are currently filtered by default. In the future, this will be handled based on user settings.)
- `view post <post_id>`: This command shows the post specified by the id argument, plus the first 20 comments.
- `view notifications (optional_arg)`: Load 15 notifications. If the notification is for a post or reply, you can use the shown id to open that post. Supports the same optional arguments as 'view feed'
- `view notes`: Loads 10 notes from your journal.
- `view bookmarks`: Load 10 bookmarks. Due to current API limitations, only bookmarked posts can be displayed, but not bookmarked replies.
- `view profile <username>`: Displays a simplified version of that users profile, as well as their pinned post if they have one. Use 'me' as the username to see your own profile.
- `write post`: Opens your default text editor (or if you have non, nano (use ctrl+s, ctrl+x to exit)) and lets you write a post. Be aware that it might fail to post, so don't invest too much effort into it without copying the contents elsewhere before saving and closing the editor. After closing the editor, you'll have a chance to choose topics for the post.
- `write reply <target_id>`: Write a reply to the post or reply whose id you gave. Will ask for final confirmation before posting.
- `write note`: Same as 'write post', but your writing is put in your journal instead.
- `edit note <note_id`: Opens a note in your default text editor (if none, nano/notepad) and lets you edit it.
- `publish <note_id>`: Posts a note to the feed, making it visible to other users.
- `edit config`: This lets you edit the client's config file. If you set 'stay logged in' to true, the client will save your refresh token and you will remain logged in across sessions. The config file should be in your .config/ or Library/Application Support/ directories, depending on whether you use linux or apple.
- `bookmark <target_id`: Bookmarks the post or reply whose id was given as an argument.
- `delete <target_id>`: This command deletes replies, posts, notes or bookmarks. The type of the target doesn't need to be specified, if will be inferred from the id. You will be asked to confirm intent to delete.
- `help`: Prints instructions to the console.
- `logout`: Log out of your account and exit the client. You will need to enter your email and password again the next time.
- `exit`: Exit without logging out



## Limitations

~~The client doesn't work on Windows, because it uses traits of the Linux terminal to format the output & edit documents.~~. The client does not have full functionality on windows. You can browse the feed with minimal visual artifacts, but it won't automatically format the color like on mac/linux, meaning you would have to set that manually in the terminal settings.
 If you run Windows, I can quickly reccomend two ways of circumventing these limitations: Install [WSL](https://learn.microsoft.com/en-us/windows/wsl/install), an official Linux subsystem for Windows, or use a different client. From conversation with the dev, I know that cyberspace user @Ragnar's TUI client works just fine on windows due to having a different technological foundation. You can find it [here](https://github.com/ArmadilloBrillo/cyber-tui).

The client doesn't support pure keyboard navigation, as people will still need to use a mouse/trackpad to scroll up and down the terminal output history to browse e.g. 'view feed' output. Supposedly this can be circumvented with `Fn + ↑ / ↓` on mac and `Shift + PageUp / PageDown` on Windows/Linux, but if at all, that only works on some versions. Personally, I managed to scroll up and down in windows 10 Powershell via `Ctrl + ↑ / ↓`, but the same didn't work in WSL Ubuntu.
