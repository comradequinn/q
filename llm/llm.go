package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/comradequinn/gen/llm/internal/resource"
	"github.com/comradequinn/gen/llm/internal/schema"
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
		DebugPrintf   func(msg string, args ...any)
	}
	User struct {
		Name        string
		Location    string
		Description string
	}
	Prompt struct {
		History   []Message
		Text      string
		Files     []string
		Schema    string
		Grounding bool
	}
	FileReference struct {
		URI      string `json:"uri"`
		MIMEType string `json:"mimeType"`
		Label    string `json:"label"`
	}
	Response struct {
		Tokens int
		Text   string
		Files  []FileReference
	}
	Role    string
	Message struct {
		Role  Role            `json:"role"`
		Text  string          `json:"text"`
		Files []FileReference `json:"files,omitempty"`
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
		if len(message.Files) > 0 {
			for _, fileReference := range message.Files {
				content.Parts = append(content.Parts, schema.Part{
					File: &schema.FileData{URI: fileReference.URI, MIMEType: fileReference.MIMEType},
				})
			}
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
		resourceRefs []resource.Reference
		err          error
	)

	if len(prompt.Files) > 0 {
		for _, f := range prompt.Files {
			resourceRef, err := resource.Upload(resource.UploadRequest{
				URL:  cfg.UploadURL,
				Key:  cfg.APIKey,
				File: f,
			}, cfg.DebugPrintf)

			if err != nil {
				return Response{}, fmt.Errorf("unable to upload file '%v' to gemini api. %v", f, err)
			}
			content.Parts = append(content.Parts, schema.Part{File: &schema.FileData{URI: resourceRef.URI, MIMEType: resourceRef.MIMEType}})
			resourceRefs = append(resourceRefs, resourceRef)
		}
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

	generationConfig.ResponseMimeType = "text/plain"

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
	cfg.DebugPrintf("sending generate request", "type", "generate_request", "url", url, "request", string(request.Bytes()))

	rs, err := http.Post(url, "application/json", &request)

	if err != nil {
		return Response{}, fmt.Errorf("unable to send request to llm api. %w", err)
	}

	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)

	if err != nil {
		return Response{}, fmt.Errorf("unable to read response body. %w", err)
	}

	cfg.DebugPrintf("received generate response", "type", "generate_response", "status", rs.Status, "request", string(body))

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

	cfg.DebugPrintf("token count value reported", "type", "report", "token_count", response.UsageMetadata.TotalTokenCount)

	files := make([]FileReference, 0, len(resourceRefs))

	for _, resourceRef := range resourceRefs {
		files = append(files, FileReference{
			URI:      resourceRef.URI,
			MIMEType: resourceRef.MIMEType,
			Label:    resourceRef.Label,
		})
	}

	return Response{
		Tokens: response.UsageMetadata.TotalTokenCount,
		Text:   sb.String(),
		Files:  files,
	}, nil
}
