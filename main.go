package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/comradequinn/q/cfg"
	"github.com/comradequinn/q/cli"
	"github.com/comradequinn/q/llm"
	"github.com/comradequinn/q/session"
)

const (
	app = "q"
)

var (
	commit = "dev"
	tag    = "none"
)

func main() {
	log.SetOutput(os.Stderr)
	homeDir, _ := os.UserHomeDir()

	version := flag.Bool("version", false, "print the version")
	versionShort := flag.Bool("v", false, "print the version")
	debug := flag.Bool("debug", false, "enable debug output")
	appDir := flag.String("app-dir", path.Join(homeDir, "."+app), fmt.Sprintf("location of the %v app (directory", app))
	configure := flag.Bool("config", false, "reset or initialise the configuration")
	model := flag.String("model", llm.ModelGeminiPro, "the model to use")
	tokens := flag.Int("tokens", 10000, "the maximum number of tokens to allow in a response")
	temperature := flag.Float64("temp", 0.1, "the temperature setting for the model")
	newSession := flag.Bool("new", false, "save any existing session and start a new one (also -n)")
	newSessionShort := flag.Bool("n", false, "save any existing session and start a new one (also --new)")
	listSessions := flag.Bool("list", false, "list all saved sessions by id (also -l)")
	listSessionsShort := flag.Bool("l", false, "list all saved sessions by id (also --list)")
	restoreSession := flag.Int("restore", 0, "the session id to restore (also -r)")
	restoreSessionShort := flag.Int("r", 0, "the session id to restore (also --restore)")
	deleteSession := flag.Int("delete", 0, "the session id to delete")
	deleteAllSessions := flag.Bool("delete-all", false, "delete all session data")
	apiURL := flag.String("url", "https://generativelanguage.googleapis.com/v1beta/models/%v:generateContent?key=%v", "the url for the gemini api. it must expose two placeholders; one for the model and a second for the api key")
	systemPrompt := flag.String("system-prompt", "Your responses are printed to a linux terminal. You will ensure those responses are concise and easily rendered in a linux terminal. "+
		"You will not use markdown syntax in your responses as this is not rendered well in terminal output. However you may use clear, plain text formatting that can be read easily and immediately by a human, "+
		"such as using dashes for list delimiters. All answers should be factually correct and you should take caution regarding hallucinations. You should only answer the specific question given; do not proactively "+
		"include additional information that is not directly relevant to the question. ", "the base system prompt to use")

	flag.Parse()

	if *debug {
		llm.LogPrintf = log.Printf
	}

	config, err := cfg.Read(*appDir)
	if err != nil {
		fmt.Printf("unable to read config file. %v", err)
	}

	if *configure || config.Credentials.APIKey == "" {
		cli.Configure(&config)
		cfg.Save(config, *appDir)

		if *configure {
			os.Exit(0)
		}
	}

	printfFatal := func(format string, v ...any) {
		fmt.Printf(format+"\n", v...)
		os.Exit(1)
	}

	switch {
	case *version || *versionShort:
		fmt.Printf("%v %v %v\n", app, tag, commit)
		os.Exit(0)
	case *newSession || *newSessionShort:
		session.Stash(*appDir)
	case *restoreSession > 0 || *restoreSessionShort > 0:
		sessionID := *restoreSession + *restoreSessionShort
		if err := session.Restore(*appDir, sessionID); err != nil {
			printfFatal("unable to restore session. %v", err)
		}
		os.Exit(0)
	case *deleteSession > 0:
		if err := session.Delete(*appDir, *deleteSession); err != nil {
			printfFatal("unable to delete session. %v", err)
		}
		os.Exit(0)
	case *deleteAllSessions:
		if err := session.DeleteAll(*appDir); err != nil {
			printfFatal("unable to delete sessions. %v", err)
		}
		os.Exit(0)
	case *listSessions || *listSessionsShort:
		records, err := session.List(*appDir)
		if err != nil {
			printfFatal("unable to list history. %v", err)
		}
		cli.ListSessions(records)
		os.Exit(0)
	}

	if len(flag.Args()) != 1 {
		printfFatal("a single prompt is required")
	}
	prompt := flag.Arg(0)

	messages, err := session.Read(*appDir)
	if err != nil {
		printfFatal("unable to read history. %v", err)
	}

	llmDone, spinnerDone := cli.Spin()

	rs, err := llm.Generate(
		llm.Config{
			APIKey:        config.Credentials.APIKey,
			APIURL:        *apiURL,
			SystemPrompt:  *systemPrompt,
			ResponseStyle: config.Preferences.ResponseStyle,
			User: llm.User{
				Name:       config.User.Name,
				Location:   config.User.Location,
				Sex:        config.User.Sex,
				Age:        config.User.Age,
				Occupation: config.User.Occupation,
			},
		},
		llm.Prompt{
			Model:       llm.Model(*model),
			MaxTokens:   *tokens,
			Temperature: *temperature,
			Text:        prompt,
			History:     messages,
		})

	if err != nil {
		printfFatal("error with llm api. %v", err)
	}

	if err := session.Write(*appDir, session.Entry{
		Prompt:   prompt,
		Response: rs.Text,
	}); err != nil {
		printfFatal("unable to update session. %v", err)
	}

	llmDone <- struct{}{}
	<-spinnerDone

	fmt.Printf("%v\n", rs.Text)
}
