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
	"github.com/comradequinn/q/schema"
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
	printfFatal := func(format string, v ...any) {
		fmt.Printf(format+"\n", v...)
		os.Exit(1)
	}

	log.SetOutput(os.Stderr)

	homeDir, _ := os.UserHomeDir()
	version := flag.Bool("version", false, "print the version")
	versionShort := flag.Bool("v", false, "print the version")
	script := flag.Bool("script", false, "supress activity indicators, such as spinners, to better support piping stdout into other utils when scripting")
	scriptShort := flag.Bool("s", false, "supress activity indicators, such as spinners, to better support piping stdout into other utils when scripting")
	grounding := flag.Bool("grounding", true, "enable grounding with search")
	schemaDefinition := flag.String("schema", "", "a schema that defines the required response format. either in the form `name:type:[description],...n` or as a json-form open-api schema. grounding with search must be disabled to use a schema")
	debug := flag.Bool("debug", false, "enable debug output")
	appDir := flag.String("app-dir", path.Join(homeDir, "."+app), fmt.Sprintf("location of the %v app (directory", app))
	configure := flag.Bool("config", false, "reset or initialise the configuration")
	model := flag.String("model", "", "the specific model to use")
	flashModel := flag.Bool("flash", false, fmt.Sprintf("use the cheaper %v model", llm.Models.Flash))
	maxTokens := flag.Int("max-tokens", 10000, "the maximum number of tokens to allow in a response")
	temperature := flag.Float64("temperature", 0.2, "the temperature setting for the model")
	topP := flag.Float64("top-p", 0.2, "the top-p setting for the model")
	apiURL := flag.String("api-url", "https://generativelanguage.googleapis.com/v1beta/models/%v:generateContent?key=%v", "the url for the gemini api. it must expose two placeholders; one for the model and a second for the api key")
	uploadURL := flag.String("upload-url", "https://generativelanguage.googleapis.com/upload/v1beta/files?key=%v", "the url for the gemini api file upload url. it must expose a placeholder for the api key")
	systemPrompt := flag.String("system-prompt", "Your responses are printed to a linux terminal. You will ensure those responses are concise and easily rendered in a linux terminal. "+
		"You will not use markdown syntax in your responses as this is not rendered well in terminal output. However you may use clear, plain text formatting that can be read easily and immediately by a human, "+
		"such as using dashes for list delimiters. All answers should be factually correct and you should take caution regarding hallucinations. You should only answer the specific question given; do not proactively "+
		"include additional information that is not directly relevant to the question. ", "the base system prompt to use")
	file := flag.String("file", "", "the path to a file to include in the prompt")
	fileShort := flag.String("f", "", "the path to a file to include in the prompt")
	newSession := flag.Bool("new", false, "save any existing session and start a new one (also -n)")
	newSessionShort := flag.Bool("n", false, "save any existing session and start a new one (also --new)")
	listSessions := flag.Bool("list", false, "list all saved sessions by id (also -l)")
	listSessionsShort := flag.Bool("l", false, "list all saved sessions by id (also --list)")
	restoreSession := flag.Int("restore", 0, "the session id to restore (also -r)")
	restoreSessionShort := flag.Int("r", 0, "the session id to restore (also --restore)")
	deleteSession := flag.Int("delete", 0, "the session id to delete")
	deleteSessionShort := flag.Int("d", 0, "the session id to delete")
	deleteAllSessions := flag.Bool("delete-all", false, "delete all session data")

	flag.Parse()

	config, err := cfg.Read(*appDir)
	if err != nil {
		printfFatal("unable to read config. %v", err)
	}

	{ // non-prompt commands
		switch {
		case *version || *versionShort:
			fmt.Printf("%v %v %v (pro-model: %v, flash-model: %v)\n", app, tag, commit, llm.Models.Pro, llm.Models.Flash)
			os.Exit(0)
		case *configure:
			cli.Configure(&config)
			cfg.Save(config.User, config.Preferences, *appDir)
			os.Exit(0)
		case *newSession || *newSessionShort:
			session.Stash(*appDir)
		case *restoreSession > 0 || *restoreSessionShort > 0:
			sessionID := *restoreSession + *restoreSessionShort
			if err := session.Restore(*appDir, sessionID); err != nil {
				printfFatal("unable to restore session. %v", err)
			}
			os.Exit(0)
		case *deleteSession > 0 || *deleteSessionShort > 0:
			sessionID := *deleteSession + *deleteSessionShort
			if err := session.Delete(*appDir, sessionID); err != nil {
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
	}

	if len(flag.Args()) != 1 {
		printfFatal("a single prompt is required")
	}
	prompt := flag.Arg(0)

	var stopSpinner = func() {}
	{
		if !*script && !*scriptShort {
			stopSpinner = cli.Spin()
		}
	}

	schema, err := schema.Build(*schemaDefinition)
	if err != nil {
		printfFatal("invalid schema definition. %v", err)
	}

	messages, err := session.Read(*appDir)
	if err != nil {
		printfFatal("unable to read history. %v", err)
	}

	useModel := *model
	{
		if useModel == "" {
			useModel = llm.Models.Pro
		}
		if *flashModel {
			useModel = llm.Models.Flash
		}
	}

	rs, err := llm.Generate(
		llm.Config{
			APIKey:        config.Credentials.APIKey,
			APIURL:        *apiURL,
			UploadURL:     *uploadURL,
			SystemPrompt:  *systemPrompt,
			ResponseStyle: config.Preferences.ResponseStyle,
			Model:         useModel,
			MaxTokens:     *maxTokens,
			Temperature:   *temperature,
			TopP:          *topP,
			User: llm.User{
				Name:        config.User.Name,
				Location:    config.User.Location,
				Description: config.User.Description,
			},
			DebugPrintf: func() func(string, ...any) {
				if !*debug {
					return func(string, ...any) {}
				}
				return log.Printf
			}(),
		},
		llm.Prompt{
			Text:      prompt,
			File:      *file + *fileShort,
			History:   messages,
			Schema:    schema,
			Grounding: *grounding,
		})

	if err != nil {
		printfFatal("error with llm api. %v", err)
	}

	if err := session.Write(*appDir, session.Entry{
		Prompt:       prompt,
		Response:     rs.Text,
		FileURI:      rs.File.URI,
		FileMIMEType: rs.File.MIMEType,
	}); err != nil {
		printfFatal("unable to update session. %v", err)
	}

	stopSpinner()

	fmt.Printf("%v\n", rs.Text)
}
