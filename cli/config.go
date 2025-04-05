package cli

import (
	"fmt"

	"github.com/comradequinn/q/cfg"
)

// Configure populates the specified configuration struct with data requested from the user
func Configure(config *cfg.Config) {
	config.Credentials.APIKey = read("Enter your Gemini API key (available from https://aistudio.google.com/apikey)", config.Credentials.APIKey)
	config.User.Name = read("Enter your name", config.User.Name)
	config.User.Location = read("Enter your location", config.User.Location)
	config.User.Occupation = read("Enter your occupation", config.User.Occupation)
	config.User.Sex = read("Enter your sex", config.User.Sex)
	config.User.Age = read("Enter your age", config.User.Age)
	config.Preferences.ResponseStyle = read("Add any desired response style (or enter 'none')", config.Preferences.ResponseStyle)

	if config.Preferences.ResponseStyle == "none" {
		config.Preferences.ResponseStyle = ""
	}
}

func read(prompt, current string) string {
	value := current

	for {
		writer(prompt)
		if current != "" {
			writer(fmt.Sprintf(" (default: '%v')", current))
		}
		writer(":\n")

		scanner.Scan()

		if scanner.Text() != "" {
			value = scanner.Text()
		}

		if value == "" {
			continue
		}
		break
	}
	writer("\n")

	return value
}
