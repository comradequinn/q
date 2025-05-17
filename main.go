package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/comradequinn/gen/cfg"
	"github.com/comradequinn/gen/cli"
	"github.com/comradequinn/gen/llm"
	"github.com/comradequinn/gen/schema"
	"github.com/comradequinn/gen/session"
)

const (
	app = "gen"
)

var (
	commit = "dev"
	tag    = "none"
)

func main() {
	checkFatalf := func(condition bool, format string, v ...any) {
		if !condition {
			return
		}
		fmt.Printf(format+"\n", v...)
		os.Exit(1)
	}

	homeDir, _ := os.UserHomeDir()
	version := flag.Bool("version", false, "print the version")
	versionShort := flag.Bool("v", false, "shortform of --version")
	script := flag.Bool("script", false, "supress activity indicators, such as spinners, to better support piping stdout into other utils when scripting")
	scriptShort := flag.Bool("s", false, "shortform of --script")
	disableGrounding := flag.Bool("no-grounding", false, "disable grounding with search")
	schemaDefinition := flag.String("schema", "", "a schema that defines the required response format. either in the form `name:type:[description],...n` or as a json-form open-api schema. grounding with search must be disabled to use a schema")
	debug := flag.Bool("debug", false, "enable debug output")
	stats := flag.Bool("stats", false, "print count of tokens used")
	appDir := flag.String("app-dir", path.Join(homeDir, "."+app), fmt.Sprintf("location of the %v app (directory", app))
	configure := flag.Bool("config", false, "reset or initialise the configuration")
	model := flag.String("model", "", "the specific model to use")
	flashModel := flag.Bool("flash", false, fmt.Sprintf("use the cheaper %v model", llm.Models.Flash))
	maxTokens := flag.Int("max-tokens", 10000, "the maximum number of tokens to allow in a response")
	temperature := flag.Float64("temperature", 0.2, "the temperature setting for the model")
	topP := flag.Float64("top-p", 0.2, "the top-p setting for the model")
	apiURL := flag.String("api-url", "https://generativelanguage.googleapis.com/v1beta/models/%v:generateContent?key=%v", "the url for the gemini api. it must expose two placeholders; one for the model and a second for the api key")
	uploadURL := flag.String("upload-url", "https://generativelanguage.googleapis.com/upload/v1beta/files?key=%v", "the url for the gemini api file upload url. it must expose a placeholder for the api key")
	systemPrompt := flag.String("system-prompt",
		fmt.Sprintf("You are a command line assistant utility named '%v' running in a terminal on the OS '%v'. Factor that into the format and content of your responses and always ensure they are concise and "+
			"easily rendered in such a terminal. You do not use complex markdown syntax in your responses as this is not rendered well in terminal output. You do use clear, plain text formatting that can be easily read and "+
			"by a human; such as using dashes for list delimiters. You always ensure that, to the extent that you are reasonably able, that your answers are factually correct and you take caution regarding hallucinations. "+
			"You only answer the specific question given and do not proactively include additional information that is not directly relevant to that question. ", app, runtime.GOOS),
		"the base system prompt to use")
	file := flag.String("files", "", "a comma separated list of files to attach to the prompt")
	fileShort := flag.String("f", "", "shortform of --files")
	newSession := flag.Bool("new", false, "save any existing session and start a new one")
	newSessionShort := flag.Bool("n", false, "shortform of --new")
	listSessions := flag.Bool("list", false, "list all sessions by id")
	listSessionsShort := flag.Bool("l", false, "shortform of --list")
	restoreSession := flag.Int("restore", 0, "the session id to restore")
	restoreSessionShort := flag.Int("r", 0, "shortform of --restore")
	deleteSession := flag.Int("delete", 0, "the session id to delete")
	deleteSessionShort := flag.Int("d", 0, "shortform of --delete")
	deleteAllSessions := flag.Bool("delete-all", false, "delete all session data")

	flag.Parse()

	logLevel := slog.LevelInfo

	if *debug {
		logLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	scriptMode := *script || *scriptShort

	config, err := cfg.Read(*appDir)
	checkFatalf(err != nil, "unable to read config. %v", err)

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
			checkFatalf(session.Restore(*appDir, sessionID) != nil, "unable to restore session. %v", err)
			os.Exit(0)
		case *deleteSession > 0 || *deleteSessionShort > 0:
			sessionID := *deleteSession + *deleteSessionShort
			checkFatalf(session.Delete(*appDir, sessionID) != nil, "unable to delete session. %v", err)
			os.Exit(0)
		case *deleteAllSessions:
			checkFatalf(session.DeleteAll(*appDir) != nil, "unable to delete sessions. %v", err)
			os.Exit(0)
		case *listSessions || *listSessionsShort:
			records, err := session.List(*appDir)
			checkFatalf(err != nil, "unable to list history. %v", err)
			cli.ListSessions(records)
			os.Exit(0)
		}
	}

	checkFatalf(len(flag.Args()) != 1, "a single prompt is required")
	prompt := flag.Arg(0)

	var stopSpinner = func() {}
	{
		if !scriptMode {
			stopSpinner = cli.Spin()
		}
	}

	schema, err := schema.Build(*schemaDefinition)
	checkFatalf(err != nil, "invalid schema definition. %v", err)

	messages, err := session.Read(*appDir)
	checkFatalf(err != nil, "unable to read history. %v", err)

	useModel := *model
	{
		if useModel == "" {
			useModel = llm.Models.Pro
		}
		if *flashModel {
			useModel = llm.Models.Flash
		}
	}

	files := []string{}
	{
		if filePattern := *file + *fileShort; filePattern != "" {
			files = strings.Split(filePattern, ",")
			for i := range files {
				files[i] = strings.TrimSpace(files[i])
			}
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
			Grounding:     !*disableGrounding,
			User: llm.User{
				Name:        config.User.Name,
				Location:    config.User.Location,
				Description: config.User.Description,
			},
			DebugPrintf: slog.Debug,
		},
		llm.Prompt{
			Text:    prompt,
			Files:   files,
			History: messages,
			Schema:  schema,
		})

	checkFatalf(err != nil, "error with llm api. %v", err)

	checkFatalf(session.Write(*appDir, session.Entry{
		Prompt:   prompt,
		Response: rs.Text,
		Files:    rs.Files,
	}) != nil, "unable to update session. %v", err)

	stopSpinner()

	fmt.Printf("%v\n\n", rs.Text)

	if *stats {
		_ = json.NewEncoder(os.Stderr).Encode(map[string]map[string]string{
			"stats": {
				"systemPromptBytes": fmt.Sprintf("%v", len(*systemPrompt)),
				"promptBytes":       fmt.Sprintf("%v", len(prompt)),
				"responseBytes":     fmt.Sprintf("%v", len(rs.Text)),
				"tokens":            fmt.Sprintf("%v", rs.Tokens),
				"files":             fmt.Sprintf("%v", len(rs.Files)),
			},
		})
	}
}
