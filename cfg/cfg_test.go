package cfg_test

import (
	"os"
	"testing"

	"github.com/comradequinn/q/cfg"
)

func TestConfig(t *testing.T) {
	testDir := "./test"

	defer func() {
		if _, err := os.Stat(testDir); err == nil {
			if err := os.RemoveAll(testDir); err != nil {
				t.Fatalf("expected no error removing test file. got %v", err)
			}
		}
	}()

	expectedCfg := cfg.Config{
		Credentials: cfg.Credentials{
			APIKey: "test-api-key",
		},
		User: cfg.User{
			Location:   "test-location",
			Name:       "test-name",
			Occupation: "test-occupation",
			Age:        "test-age",
			Sex:        "test-sex",
		},
		Preferences: cfg.Preferences{
			ResponseStyle: "test-response-style",
		},
	}

	if err := cfg.Save(expectedCfg, testDir); err != nil {
		t.Fatalf("expected no error saving config. got %v", err)
	}

	actualCfg, err := cfg.Read(testDir)

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

	if actualCfg.User.Occupation != expectedCfg.User.Occupation {
		t.Fatalf("expected occupation to be %v. got %v", expectedCfg.User.Occupation, actualCfg.User.Occupation)
	}

	if actualCfg.User.Age != expectedCfg.User.Age {
		t.Fatalf("expected age to be %v. got %v", expectedCfg.User.Age, actualCfg.User.Age)
	}

	if actualCfg.User.Sex != expectedCfg.User.Sex {
		t.Fatalf("expected sex to be %v. got %v", expectedCfg.User.Sex, actualCfg.User.Sex)
	}

	if actualCfg.Preferences.ResponseStyle != expectedCfg.Preferences.ResponseStyle {
		t.Fatalf("expected response style to be %v. got %v", expectedCfg.Preferences.ResponseStyle, actualCfg.Preferences.ResponseStyle)
	}
}
