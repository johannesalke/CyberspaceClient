package client

import (
	"encoding/json"
	"errors"
	"fmt"
	//"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type APIClient struct {
	Client            *http.Client
	Tokens            AuthTokens
	ApiUrl            string
	UserID            string
	Username          string
	PostCache         map[string]Post         // key:PostID
	ReplyCache        map[string]Reply        // key: ReplyID
	NotificationCache map[string]Notification // key:PostID
	NoteCache         map[string]Note
	BookmarkCache     map[string]Bookmark
	Cursors           map[string]string // key: whatever you want
	LastStatusCode    int
	Config            Config
}

const CyberspaceApiUrl = "https://api.cyberspace.online/v1"

func InitAPIClient() APIClient {
	return APIClient{
		ApiUrl:            CyberspaceApiUrl,
		Client:            &http.Client{},
		PostCache:         make(map[string]Post),
		NotificationCache: make(map[string]Notification),
		ReplyCache:        make(map[string]Reply),
		NoteCache:         make(map[string]Note),
		BookmarkCache:     make(map[string]Bookmark),
		Cursors:           make(map[string]string),
	}
}

type Config struct {
	StoredValues ConfigStorage  `json:"stored_values"`
	Settings     ConfigSettings `json:"settings"`
}

type ConfigStorage struct {
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type ConfigSettings struct {
	AutoResize   bool                `json:"auto_resize"`
	StayLoggedIn bool                `json:"stay_logged_in"`
	KeyBindings  map[string][]string `json:"keybindings"`
	Theme        string              `json:"theme"`
}

// DefaultKeyBindings returns a fresh copy of the TUI's keyboard bindings.
// Each action can be replaced with any list of Bubble Tea key names in the
// user's config file. An empty list intentionally disables that action.
func DefaultKeyBindings() map[string][]string {
	return map[string][]string{
		"quit":                    {"q", "ctrl+c"},
		"help":                    {"?"},
		"close_help":              {"esc"},
		"refresh":                 {"r"},
		"next_page":               {"n", "right"},
		"scroll_up":               {"up", "k"},
		"scroll_down":             {"down", "j"},
		"page_up":                 {"pgup", "ctrl+u"},
		"page_down":               {"pgdown", "ctrl+d", "space"},
		"top":                     {"home", "g"},
		"bottom":                  {"end", "G"},
		"page_feed":               {"1"},
		"page_bookmarks":          {"2"},
		"page_notifications":      {"3"},
		"page_journal":            {"4"},
		"page_profile":            {"5"},
		"page_mail":               {"6"},
		"page_jukebox":            {"7"},
		"compose_post":            {"c"},
		"submit_post":             {"ctrl+s"},
		"confirm_post":            {"enter", "y"},
		"cancel_compose":          {"esc"},
		"switch_theme":            {"t"},
		"select_next":             {"tab", "j"},
		"select_previous":         {"shift+tab", "k"},
		"open_post":               {"enter"},
		"back":                    {"esc"},
		"toggle_bookmark":         {"b"},
		"reply_to_post":           {"r"},
		"jukebox_select_next":     {"down", "j"},
		"jukebox_select_previous": {"up", "k"},
		"jukebox_play":            {"enter", "space"},
		"jukebox_pause":           {"p"},
		"jukebox_next":            {"right"},
		"jukebox_previous":        {"left"},
		"jukebox_stop":            {"x"},
		"jukebox_page_next":       {"n", "pgdown"},
		"jukebox_page_previous":   {"pgup"},
	}
}

// ResolveKeyBindings applies user overrides to the defaults without mutating
// either map. Keeping this logic in the client package lets every interface
// share the same user configuration.
func ResolveKeyBindings(overrides map[string][]string) map[string][]string {
	bindings := DefaultKeyBindings()
	for action, keys := range overrides {
		bindings[action] = append([]string(nil), keys...)
	}
	return bindings
}

func ensureKeyBindings(settings *ConfigSettings) bool {
	if settings.KeyBindings == nil {
		settings.KeyBindings = DefaultKeyBindings()
		return true
	}

	changed := false
	for action, keys := range DefaultKeyBindings() {
		if _, exists := settings.KeyBindings[action]; !exists {
			settings.KeyBindings[action] = keys
			changed = true
		}
	}
	return changed
}

func ensureTheme(settings *ConfigSettings) bool {
	if settings.Theme != "" {
		return false
	}
	settings.Theme = "amber"
	return true
}

//Missing: Follows,

//Incomplete: Users(Profile update)

func GetConfig() (config Config) {
	cfg, err := GetConfigDir()
	if err != nil {
		fmt.Printf("Critical error: Couldn't retrieve config path. %s", err)
	}
	cfg_file_path := filepath.Join(cfg, "config.json")

	//If config doesn't exist, create it with default values.
	if _, err := os.Stat(cfg_file_path); errors.Is(err, os.ErrNotExist) {
		InitConfig(cfg_file_path)
	}
	config, err = readConfig(cfg_file_path)
	if err != nil {
		fmt.Printf("Couldn't retrieve config. %s", err)
	}
	keyBindingsChanged := ensureKeyBindings(&config.Settings)
	themeChanged := ensureTheme(&config.Settings)
	if keyBindingsChanged || themeChanged {
		if err := writeConfig(cfg_file_path, config); err != nil {
			fmt.Printf("Couldn't save default TUI settings. %s", err)
		}
	}
	return config
}

func (c *APIClient) UpdateConfig() (config Config) {
	cfg, err := GetConfigDir()
	if err != nil {
		fmt.Printf("Critical error: Couldn't retrieve config path. %s", err)
	}
	cfgPath := filepath.Join(cfg, "config.json")
	file, err := os.ReadFile(cfgPath)
	if err != nil {
		fmt.Printf("Error unmarshalling config json: %s", err)

	}
	tmpFile, err := os.CreateTemp("", "config-*.txt")
	if err != nil {
		fmt.Printf("Error unmarshalling config json: %s", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(string(file))
	if err != nil {
		fmt.Printf("Error writing to temp: %s", err)
	}
	tmpFile.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano" // fallback
	}
	if runtime.GOOS == "windows" {
		editor = "notepad"
	}

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error unmarshalling config json: %s", err)
	}

	file, err = os.ReadFile(tmpFile.Name())
	if err != nil {
		fmt.Printf("Error unmarshalling config json: %s", err)
	}

	err = json.Unmarshal(file, &config)
	if err != nil {

		fmt.Printf("Error unmarshalling config json: %s", err)
	}
	if config.Settings.StayLoggedIn == true {
		config.StoredValues.RefreshToken = c.Tokens.RefreshToken
	}
	file, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error unmarshalling config json: %s", err)
	}

	err = os.WriteFile(cfgPath, file, 0644)
	if err != nil {
		fmt.Printf("Error writing config: %s", err)
	}

	err = json.Unmarshal(file, &config)
	if err != nil {

		fmt.Printf("Error unmarshalling config json: %s", err)
	}
	return config

}

func (c *APIClient) SaveConfig(config Config) error {
	cfgDir, err := GetConfigDir()
	if err != nil {
		return fmt.Errorf("Critical error: Couldn't retrieve config path. %s", err)
	}
	cfgPath := filepath.Join(cfgDir, "config.json")

	/*if config.StayLoggedIn {
		config.StoredValues.RefreshToken = c.Tokens.RefreshToken
	}*/

	file, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("Error unmarshalling config json: %s", err)
	}

	err = os.WriteFile(cfgPath, file, 0644)
	if err != nil {
		return fmt.Errorf("Error writing config: %s", err)
	}
	return nil
}

// //////| Config helper functions |///////////////////
func readConfig(cfgPath string) (Config, error) {
	file, err := os.ReadFile(cfgPath)
	if err != nil {
		return Config{}, fmt.Errorf("Error unmarshalling config json: %s", err)
	}
	var config Config
	//fmt.Println(string(file))

	err = json.Unmarshal(file, &config)
	if err != nil {

		return Config{}, fmt.Errorf("Error unmarshalling config json: %s", err)
	}
	return config, nil
}

func InitConfig(cfgPath string) error {
	config := Config{

		StoredValues: ConfigStorage{

			RefreshToken: "",
		},
		Settings: ConfigSettings{
			AutoResize:   false,
			StayLoggedIn: true,
			KeyBindings:  DefaultKeyBindings(),
			Theme:        "amber",
		},
	}
	return writeConfig(cfgPath, config)
}

func writeConfig(cfgPath string, config Config) error {
	configJson, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("Error marshalling config json: %s", err)
	}
	err = os.WriteFile(cfgPath, configJson, 0644)
	if err != nil {
		return fmt.Errorf("Error writing config: %s", err)
	}
	return nil

}

func GetConfigDir() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		configDir = os.Getenv("APPDATA")
	case "darwin":
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, "Library", "Application Support")
	default: // linux and others
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig != "" {
			configDir = xdgConfig
		} else {
			home, _ := os.UserHomeDir()
			configDir = filepath.Join(home, ".config")
		}
	}

	appConfigDir := filepath.Join(configDir, "cyberspace_client")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", err
	}

	return appConfigDir, nil
}
