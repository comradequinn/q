package cfg

import (
	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	testDir := "./test"
	os_Getenv = func(string) string { return "test-api-key" }

	defer func() {
		if _, err := os.Stat(testDir); err == nil {
			if err := os.RemoveAll(testDir); err != nil {
				t.Fatalf("expected no error removing test file. got %v", err)
			}
		}
	}()

	expectedCfg := Config{
		Credentials: Credentials{
			APIKey: "test-api-key",
		},
		User: User{
			Location:    "test-location",
			Name:        "test-name",
			Description: "test-description",
		},
		Preferences: Preferences{
			ResponseStyle: "test-response-style",
		},
	}

	if err := Save(expectedCfg.User, expectedCfg.Preferences, testDir); err != nil {
		t.Fatalf("expected no error saving config. got %v", err)
	}

	actualCfg, err := Read(testDir)

	if err != nil {
		t.Fatalf("expected no error reading config. got %v", err)
	}

	if actualCfg.Credentials.APIKey != expectedCfg.Credentials.APIKey {
		t.Fatalf("expected api key to be %v. got %v", expectedCfg.Credentials.APIKey, actualCfg.Credentials.APIKey)
	}

	if actualCfg.User.Location != expectedCfg.User.Location {
		t.Fatalf("expected location to be %v. got %v", expectedCfg.User.Location, actualCfg.User.Location)
	}

	if actualCfg.User.Name != expectedCfg.User.Name {
		t.Fatalf("expected name to be %v. got %v", expectedCfg.User.Name, actualCfg.User.Name)
	}

	if actualCfg.User.Description != expectedCfg.User.Description {
		t.Fatalf("expected description to be %v. got %v", expectedCfg.User.Description, actualCfg.User.Description)
	}

	if actualCfg.Preferences.ResponseStyle != expectedCfg.Preferences.ResponseStyle {
		t.Fatalf("expected response style to be %v. got %v", expectedCfg.Preferences.ResponseStyle, actualCfg.Preferences.ResponseStyle)
	}
}
