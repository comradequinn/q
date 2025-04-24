package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/comradequinn/q/llm/internal/resource"
	"github.com/comradequinn/q/llm/internal/schema"
)

type (
	Config struct {
		APIKey        string
		APIURL        string
		UploadURL     string
		SystemPrompt  string
		ResponseStyle string
		Model         string
		MaxTokens     int
		Temperature   float64
		TopP          float64
		User          User
		DebugPrintf   func(format string, v ...any)
	}
	User struct {
		Name        string
		Location    string
		Description string
	}
	Prompt struct {
		History   []Message
		Text      string
		File      string
		Schema    string
		Grounding bool
	}
	FileReference struct {
		URI      string
		MIMEType string
	}
	Response struct {
		Tokens int
		Text   string
		File   FileReference
	}
	Role    string
	Message struct {
		Role         Role   `json:"role,omitzero"`
		Text         string `json:"text,omitzero"`
		FileURI      string `json:"fileURI,omitzero"`
		FileMIMEType string `json:"fileMIMEType,omitzero"`
	}
)

var (
	Models = struct {
		Pro   string
		Flash string
	}{
		Pro:   "gemini-2.5-pro-preview-03-25",
		Flash: "gemini-2.5-flash-preview-04-17",
	}
)

const (
	RoleUser  = "user"
	RoleModel = "model"
)

// Generate queries the configured LLM with the specified prompt and returns the result
func Generate(cfg Config, prompt Prompt) (Response, error) {
	if cfg.Model == "" || cfg.MaxTokens == 0 || cfg.Temperature == 0 {
		return Response{}, fmt.Errorf("invalid prompt. model, maxtokens and temperature must be specified")
	}

	if prompt.Schema != "" && prompt.Grounding {
		cfg.DebugPrintf("grounding was specified but silently disabled due to the specification of a schema. the gemini api will not currently perform grounding for prompts requiring a structured response")
		prompt.Grounding = false
	}

	systemPrompt := strings.Builder{}
	systemPrompt.WriteString(cfg.SystemPrompt + ". ")
	systemPrompt.WriteString(`Your responses are printed to a linux terminal. 
		You will ensure those responses are concise and easily rendered in a linux terminal.
		You will not use markdown syntax in your responses as this is not rendered well in terminal output. 
		However you may use clear, plain text formatting that can be read easily and immediately by a human, such as using dashes for list delimiters. 
		All answers should be factually correct and you should take caution regarding hallucinations. 
		You should only answer the specific question given; do not proactively include additional information that is not directly relevant to the question. 
		`)
	systemPrompt.WriteString(fmt.Sprintf("Your responses must not exceed %v words in length. ", float64(cfg.MaxTokens)*0.75)) // rough mapping of tokens to words

	defineAttribute := func(key string, val any, unset any) string {
		if val == unset {
			return ""
		}
		return fmt.Sprintf("Consider in your responses, where it may be relevant, that the user has provided this information regarding their %v: %q", key, val) + ". "
	}

	systemPrompt.WriteString(defineAttribute("location", cfg.User.Location, ""))
	systemPrompt.WriteString(defineAttribute("name", cfg.User.Name, ""))
	systemPrompt.WriteString(defineAttribute("description", cfg.User.Description, ""))
	systemPrompt.WriteString(defineAttribute("preferred response style; note that this only refines your output and does not override any previous instruction where there is a contradiction", cfg.ResponseStyle, ""))

	contents := make([]schema.Content, 0, len(prompt.History)+1)

	for _, message := range prompt.History {
		content := schema.Content{
			Role:  string(message.Role),
			Parts: []schema.Part{{Text: message.Text}}}
		if message.FileURI != "" {
			content.Parts = append(content.Parts, schema.Part{
				File: &schema.FileData{URI: message.FileURI, MIMEType: message.FileMIMEType},
			})
		}

		contents = append(contents, content)
	}

	content := schema.Content{
		Role: RoleUser,
		Parts: []schema.Part{
			{Text: prompt.Text},
		},
	}

	var (
		resourceRef resource.Reference
		err         error
	)

	if prompt.File != "" {
		resourceRef, err = resource.Upload(resource.UploadRequest{
			URL:  cfg.UploadURL,
			Key:  cfg.APIKey,
			File: prompt.File,
		}, cfg.DebugPrintf)

		if err != nil {
			return Response{}, fmt.Errorf("unable to upload file to gemini api. %v", err)
		}

		content.Parts = append(content.Parts, schema.Part{File: &schema.FileData{URI: resourceRef.URI, MIMEType: resourceRef.MIMEType}})
	}

	contents = append(contents, content)

	tools := []schema.Tool{}

	if prompt.Grounding {
		tools = []schema.Tool{
			{GoogleSearch: &schema.GoogleSearch{}},
		}
	}

	generationConfig := schema.GenerationConfig{
		Temperature:     cfg.Temperature,
		TopP:            cfg.TopP,
		MaxOutputTokens: cfg.MaxTokens,
	}

	if prompt.Schema != "" {
		generationConfig.ResponseMimeType = "application/json"
		generationConfig.ResponseSchema = json.RawMessage(prompt.Schema)
	}

	request := bytes.Buffer{}
	if err := json.NewEncoder(&request).Encode(schema.Request{
		SystemInstruction: schema.SystemInstruction{
			Parts: []schema.Part{{Text: systemPrompt.String()}},
		},
		Contents:         contents,
		Tools:            tools,
		GenerationConfig: generationConfig,
	}); err != nil {
		return Response{}, fmt.Errorf("unable to encode llm request as json. %w", err)
	}

	url := fmt.Sprintf(cfg.APIURL, cfg.Model, cfg.APIKey)
	cfg.DebugPrintf("url=%v request=%q", url, request.Bytes())

	rs, err := http.Post(url, "application/json", &request)

	if err != nil {
		return Response{}, fmt.Errorf("unable to send request to llm api. %w", err)
	}

	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)

	if err != nil {
		return Response{}, fmt.Errorf("unable to read response body. %w", err)
	}

	cfg.DebugPrintf("response=%q", string(body))

	if rs.StatusCode != 200 {
		return Response{}, fmt.Errorf("non-200 status code returned from llm api. %s", body)
	}

	response := schema.Response{}

	if err := json.Unmarshal(body, &response); err != nil {
		return Response{}, fmt.Errorf("unable to parse response body. %w", err)
	}

	if len(response.Candidates) == 0 || response.Candidates[0].FinishReason != schema.FinishReasonStop {
		return Response{}, fmt.Errorf("no valid response candidates returned. response: %s", body)
	}

	sb := strings.Builder{}

	for _, part := range response.Candidates[0].Content.Parts {
		sb.WriteString(part.Text)
	}

	cfg.DebugPrintf("token_count=%v", response.UsageMetadata.TotalTokenCount)

	return Response{
		Tokens: response.UsageMetadata.TotalTokenCount,
		Text:   sb.String(),
		File: FileReference{
			URI:      resourceRef.URI,
			MIMEType: resourceRef.MIMEType,
		},
	}, nil
}
