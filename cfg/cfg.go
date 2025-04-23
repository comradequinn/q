package cfg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
)

type (
	Config struct {
		Credentials Credentials `json:"-"`
		User        User        `json:"user"`
		Preferences Preferences `json:"preferences"`
	}
	Credentials struct {
		APIKey string
	}
	User struct {
		Location    string `json:"location"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	Preferences struct {
		ResponseStyle string `json:"responseStyle"`
	}
)

var (
	os_Getenv = os.Getenv
)

// Read returns configuration data based on the contents environment variables and a config file
// in the specified app directory. If the file does not exist, it is created
func Read(appDir string) (Config, error) {
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return Config{}, fmt.Errorf("unable to create config file directory: %s: %w", appDir, err)
	}

	filePath := path.Join(appDir, "config")

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		return Config{}, fmt.Errorf("unable to open config file %s: %w", filePath, err)
	}

	defer file.Close()

	config := Config{}

	buffer, err := io.ReadAll(file)

	if err != nil {
		return Config{}, fmt.Errorf("unable to read config file %s: %w", filePath, err)
	}

	switch {
	case len(buffer) == 0:
		if err := Save(config.User, config.Preferences, appDir); err != nil {
			return Config{}, fmt.Errorf("unable to write new config file %s: %w", appDir, err)
		}
	default:
		if err := json.NewDecoder(bytes.NewReader(buffer)).Decode(&config); err != nil {
			return Config{}, fmt.Errorf("unable to parse config file %s: %w", filePath, err)
		}
	}

	config.Credentials.APIKey = os_Getenv("GEMINI_API_KEY")

	if config.Credentials.APIKey == "" {
		return Config{}, fmt.Errorf("unable to read the gemini-api-key from the GEMINI_API_KEY environment variable")
	}

	return config, nil
}

// Save writes the specified configuration to a config file in the specified app directory
func Save(user User, preferences Preferences, appDir string) error {
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("unable to create config file directory: %s: %w", appDir, err)
	}

	buffer := bytes.Buffer{}

	jsonEncoder := json.NewEncoder(&buffer)
	jsonEncoder.SetIndent("", "  ")

	cfg := Config{
		User:        user,
		Preferences: preferences,
	}

	if err := jsonEncoder.Encode(&cfg); err != nil {
		return fmt.Errorf("unable to encode config file %s: %w", appDir, err)
	}

	if err := os.WriteFile(path.Join(appDir, "config"), buffer.Bytes(), 0644); err != nil {
		return fmt.Errorf("unable to write config file %s: %w", appDir, err)
	}

	return nil
}
