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
	StayLoggedIn bool `json:"stay_logged_in"`

	StoredValues ConfigStorage `json:"stored_values"`
}

type ConfigStorage struct {
	Email        string `json:"email"`
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
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
	if config.StayLoggedIn == true {
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
		StayLoggedIn: false,
		StoredValues: ConfigStorage{
			Email:        "",
			RefreshToken: "",
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
