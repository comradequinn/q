package cli

import (
	"fmt"

	"github.com/comradequinn/gen/cfg"
)

// Configure populates the specified configuration struct with data requested from the user
func Configure(config *cfg.Config) {
	config.User.Name = read("Enter a name to refer to the user with", config.User.Name)
	config.User.Location = read("Enter a location relevant to the user", config.User.Location)
	config.User.Description = read("Enter a description or any information about the user that should be considered in responses", config.User.Description)
	config.Preferences.ResponseStyle = read("Add any desired response style, such as a specific tone, attitude or adopted role (enter 'none' to skip)", config.Preferences.ResponseStyle)

	if config.Preferences.ResponseStyle == "none" {
		config.Preferences.ResponseStyle = ""
	}
}

func read(prompt, defaultValue string) string {
	value := defaultValue

	for {
		writer(prompt)
		if defaultValue != "" {
			writer(fmt.Sprintf(" (default: '%v')", defaultValue))
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
