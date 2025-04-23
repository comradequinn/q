package cli

import (
	"fmt"
	"strconv"

	"github.com/comradequinn/q/cfg"
)

// Configure populates the specified configuration struct with data requested from the user
func Configure(config *cfg.Config) {
	verifyString := func(value string) string {
		return value
	}

	config.Credentials.APIKey = read("Enter a Gemini API key or enter 'env' to read it from the GEMINI_API_KEY environment variable (keys available free from https://aistudio.google.com/apikey)", config.Credentials.APIKey, verifyString)
	config.User.Name = read("Enter a name to refer to the user with", config.User.Name, verifyString)
	config.User.Location = read("Enter a location relevant to the user", config.User.Location, verifyString)
	config.User.Description = read("Enter a description or any information about the user that should be considered in responses", config.User.Description, verifyString)
	config.Preferences.ResponseStyle = read("Add any desired response style, such as a specific tone, attitude or adopted role (enter 'none' to skip)", config.Preferences.ResponseStyle, verifyString)

	if config.Preferences.ResponseStyle == "none" {
		config.Preferences.ResponseStyle = ""
	}

	verifyFloat := func(max float64) func(value string) string {
		return func(value string) string {
			f, err := strconv.ParseFloat(value, 64)

			if err != nil || f > max || f < 0 {
				return ""
			}

			return value
		}
	}

	toFloat := func(value string) float64 {
		f, _ := strconv.ParseFloat(value, 64)
		return f
	}

	temperature := read("Specify the temperature setting for the model. This can be overridden on a per-request basis. The suggested value is 0.2. The max value is 2", fmt.Sprintf("%v", config.Preferences.Temperature), verifyFloat(2))
	config.Preferences.Temperature = toFloat(temperature)

	topP := read("Specify the top-p setting for the model. This can be overridden on a per-request basis. The suggested value is 0.2. The max value is 1", fmt.Sprintf("%v", config.Preferences.TopP), verifyFloat(1))
	config.Preferences.TopP = toFloat(topP)

	verifyInt := func(value string) string {
		i, err := strconv.Atoi(value)

		if err != nil || i <= 0 {
			return ""
		}

		return value
	}

	toInt := func(value string) int {
		i, _ := strconv.Atoi(value)
		return i
	}

	if config.Preferences.MaxTokens == 0 {
		config.Preferences.MaxTokens = 10000
	}
	maxTokens := read("Specify the max tokens setting for the model. This can be overridden on a per-request basis", fmt.Sprintf("%v", config.Preferences.MaxTokens), verifyInt)
	config.Preferences.MaxTokens = toInt(maxTokens)
}

func read(prompt, defaultValue string, verifyFunc func(string) string) string {
	value := defaultValue

	for {
		writer(prompt)
		if defaultValue != "" {
			writer(fmt.Sprintf(" (default: '%v')", defaultValue))
		}
		writer(":\n")

		scanner.Scan()

		if scanner.Text() != "" {
			value = verifyFunc(scanner.Text())
		}

		if value == "" {
			continue
		}
		break
	}
	writer("\n")

	return value
}
